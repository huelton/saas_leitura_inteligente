package handler

import (
	"context"
	"regexp"
	"strings"
	"unicode"

	"github.com/huelton/leitura-inteligente/internal/ai"
	"github.com/huelton/leitura-inteligente/internal/limiter"
	"github.com/huelton/leitura-inteligente/internal/repository"
)

type mcqPrewarmTask struct {
	UserID string
	Plan   string
	BookID string
	Page   int
}

var nonWordRE = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)

func (s *Server) enqueueMCQPrewarm(task mcqPrewarmTask) {
	if s.mcqPrewarmQueue == nil {
		return
	}
	select {
	case s.mcqPrewarmQueue <- task:
	default:
		// Descarta quando fila lota para não impactar latência HTTP.
	}
}

func (s *Server) runMCQPrewarmWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-s.mcqPrewarmQueue:
			_, _ = s.ensureMCQStock(ctx, task.UserID, task.Plan, task.BookID, task.Page, targetMCQPerPage, true)
		}
	}
}

func (s *Server) ensureMCQStock(ctx context.Context, uid, plan, bookID string, page, target int, enforceLimiter bool) (int, error) {
	qRepo := repository.NewQuestionRepository(s.DB)
	cachedCount, err := qRepo.CountMCQByBookPage(ctx, bookID, page)
	if err != nil {
		return 0, err
	}
	if cachedCount >= target {
		return 0, nil
	}

	toGenerate := target - cachedCount
	if enforceLimiter {
		if err := limiter.CheckQuestions(ctx, s.DB, s.Limits, plan, uid, toGenerate); err != nil {
			return 0, err
		}
	}

	bookRepo := repository.NewBookRepository(s.DB)
	content, err := bookRepo.GetPageContent(ctx, bookID, page)
	if err != nil {
		return 0, err
	}

	items, err := s.AI.GenerateQuestionsMCQ(ctx, content, toGenerate)
	if err != nil {
		return 0, err
	}

	existing, err := qRepo.ListMCQTextsByBookPage(ctx, bookID, page)
	if err != nil {
		return 0, err
	}
	filtered := dedupeMCQ(existing, items)

	generatedNow := 0
	for _, it := range filtered {
		if strings.TrimSpace(it.Question) == "" || len(it.Options) < 2 {
			continue
		}
		_, err := qRepo.SaveMCQQuestion(ctx, bookID, page, repository.MCQRow{
			Question:   it.Question,
			Difficulty: it.Difficulty,
			Options:    it.Options,
			CorrectIdx: it.CorrectIdx,
		})
		if err != nil {
			return generatedNow, err
		}
		generatedNow++
	}

	return generatedNow, nil
}

func dedupeMCQ(existing []string, incoming []ai.MCQQuestion) []ai.MCQQuestion {
	seen := make([]string, 0, len(existing)+len(incoming))
	seen = append(seen, existing...)
	out := make([]ai.MCQQuestion, 0, len(incoming))
	for _, cand := range incoming {
		q := strings.TrimSpace(cand.Question)
		if q == "" {
			continue
		}
		dup := false
		for _, ex := range seen {
			if semanticallySimilarQuestion(q, ex) {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		seen = append(seen, q)
		out = append(out, cand)
	}
	return out
}

func semanticallySimilarQuestion(a, b string) bool {
	na := normalizeQuestion(a)
	nb := normalizeQuestion(b)
	if na == "" || nb == "" {
		return false
	}
	if na == nb {
		return true
	}
	if len(na) > 28 && len(nb) > 28 && (strings.Contains(na, nb) || strings.Contains(nb, na)) {
		return true
	}
	return tokenJaccard(na, nb) >= 0.82
}

func normalizeQuestion(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonWordRE.ReplaceAllString(s, " ")
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

func tokenJaccard(a, b string) float64 {
	ta := strings.Fields(a)
	tb := strings.Fields(b)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}
	setA := make(map[string]struct{}, len(ta))
	setB := make(map[string]struct{}, len(tb))
	for _, t := range ta {
		setA[t] = struct{}{}
	}
	for _, t := range tb {
		setB[t] = struct{}{}
	}
	inter := 0
	union := len(setA)
	for k := range setB {
		if _, ok := setA[k]; ok {
			inter++
			continue
		}
		union++
	}
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}
