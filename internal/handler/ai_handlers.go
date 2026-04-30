package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/limiter"
	"github.com/huelton/leitura-inteligente/internal/middleware"
	"github.com/huelton/leitura-inteligente/internal/repository"
)

const (
	maxQuestionsPerRequest = 5
	defaultMCQResponseSize = 5
	maxMCQResponseSize     = 30
	targetMCQPerPage       = 30
)

type pageQuestionReq struct {
	BookID string `json:"book_id" binding:"required"`
	Page   int    `json:"page" binding:"required"`
	Count  int    `json:"count"`
}

func (s *Server) GenerateQuestionsByPage(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	var req pageQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !s.assertBookOwner(c, req.BookID) {
		return
	}
	plan := middleware.CurrentPlan(c)
	if err := limiter.CheckQuestions(c.Request.Context(), s.DB, s.Limits, plan, uid, maxQuestionsPerRequest); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	bookRepo := repository.NewBookRepository(s.DB)
	content, err := bookRepo.GetPageContent(c.Request.Context(), req.BookID, req.Page)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	items, err := s.AI.GenerateQuestions(c.Request.Context(), content)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	qRepo := repository.NewQuestionRepository(s.DB)
	out := make([]gin.H, 0, len(items))
	saved := 0
	for _, it := range items {
		if it.Text == "" {
			continue
		}
		id, err := qRepo.SaveQuestion(c.Request.Context(), req.BookID, req.Page, it.Text, it.Difficulty)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		saved++
		out = append(out, gin.H{
			"id":         id,
			"question":   it.Text,
			"difficulty": it.Difficulty,
			"page":       req.Page,
			"book_id":    req.BookID,
		})
	}
	if saved > 0 {
		ur := repository.NewUsageRepository(s.DB)
		if err := ur.AddQuestions(c.Request.Context(), uid, time.Now().UTC(), saved); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"questions": out})
}

func (s *Server) GenerateQuestionsByPageMCQ(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	var req pageQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !s.assertBookOwner(c, req.BookID) {
		return
	}
	qRepo := repository.NewQuestionRepository(s.DB)
	want := req.Count
	if want <= 0 {
		want = defaultMCQResponseSize
	}
	if want > maxMCQResponseSize {
		want = maxMCQResponseSize
	}

	plan := middleware.CurrentPlan(c)
	generatedNow, err := s.ensureMCQStock(c.Request.Context(), uid, plan, req.BookID, req.Page, targetMCQPerPage, true)
	if err != nil {
		if strings.Contains(err.Error(), "limite diário") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if generatedNow > 0 {
		ur := repository.NewUsageRepository(s.DB)
		if err := ur.AddQuestions(c.Request.Context(), uid, time.Now().UTC(), generatedNow); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	rows, err := qRepo.ListRandomMCQByBookPage(c.Request.Context(), req.BookID, req.Page, want)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":          r.ID,
			"question":    r.Question,
			"difficulty":  r.Difficulty,
			"options":     r.Options,
			"correct_idx": r.CorrectIdx,
			"page":        req.Page,
			"book_id":     req.BookID,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"questions": out,
		"cache": gin.H{
			"target_per_page": targetMCQPerPage,
			"generated_now":   generatedNow,
			"served_count":    len(out),
		},
	})
}

type answerReq struct {
	QuestionID string `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

func (s *Server) SubmitAnswer(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	var req answerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	qRepo := repository.NewQuestionRepository(s.DB)
	meta, err := qRepo.GetQuestionMeta(c.Request.Context(), req.QuestionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if !s.assertBookOwner(c, meta.BookID) {
		return
	}

	score := 0
	feedback := ""
	if meta.QuestionType == "mcq" && len(meta.Options) > 0 && meta.CorrectIdx != nil {
		correct := ""
		if *meta.CorrectIdx >= 0 && *meta.CorrectIdx < len(meta.Options) {
			correct = meta.Options[*meta.CorrectIdx]
		}
		if strings.EqualFold(strings.TrimSpace(req.Answer), strings.TrimSpace(correct)) {
			score = 10
			feedback = "Resposta correta."
		} else {
			score = 3
			feedback = "Resposta incorreta. Revise a página e tente novamente."
		}
	} else {
		scoreAI, feedbackAI, err := s.AI.EvaluateAnswer(c.Request.Context(), meta.Question, req.Answer)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		score = scoreAI
		feedback = feedbackAI
	}

	if err := qRepo.SaveAnswer(c.Request.Context(), req.QuestionID, uid, req.Answer, score, feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	avg, err := qRepo.AverageScoreForBook(c.Request.Context(), uid, meta.BookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	comprehension := avg * 10.0
	if err := qRepo.UpsertReadingScore(c.Request.Context(), uid, meta.BookID, comprehension); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"score":                  score,
		"feedback":               feedback,
		"book_comprehension_pct": comprehension,
	})
}
