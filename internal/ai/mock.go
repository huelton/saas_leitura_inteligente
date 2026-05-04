package ai

import "context"

// MockService implementa Service com respostas fixas (testes de integração).
type MockService struct {
	Questions      []GeneratedQuestion
	MCQQuestions   []MCQQuestion
	EvalScore      int
	EvalFeedback   string
	SummaryPayload *ChapterSummaryPayload
	SummaryJSON    string
	Flashcards     []FlashcardPair
	Err            error
}

func (m *MockService) GenerateQuestions(_ context.Context, _ string) ([]GeneratedQuestion, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Questions, nil
}

func (m *MockService) GenerateQuestionsMCQ(_ context.Context, _ string, count int) ([]MCQQuestion, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if len(m.MCQQuestions) > 0 {
		if count > 0 && len(m.MCQQuestions) > count {
			return m.MCQQuestions[:count], nil
		}
		return m.MCQQuestions, nil
	}
	out := []MCQQuestion{
		{
			Question:   "Pergunta exemplo?",
			Difficulty: "literal",
			Options:    []string{"Opção A", "Opção B", "Opção C", "Opção D"},
			CorrectIdx: 0,
		},
	}
	if count <= 1 {
		return out, nil
	}
	for i := 1; i < count; i++ {
		out = append(out, MCQQuestion{
			Question:   "Pergunta exemplo?",
			Difficulty: "interpretacao",
			Options:    []string{"Opção A", "Opção B", "Opção C", "Opção D"},
			CorrectIdx: 0,
		})
	}
	return out, nil
}

func (m *MockService) EvaluateAnswer(_ context.Context, _, _ string) (int, string, error) {
	if m.Err != nil {
		return 0, "", m.Err
	}
	return m.EvalScore, m.EvalFeedback, nil
}

func (m *MockService) GenerateChapterSummary(_ context.Context, _ string) (*ChapterSummaryPayload, string, error) {
	if m.Err != nil {
		return nil, "", m.Err
	}
	if m.SummaryPayload != nil && m.SummaryJSON != "" {
		return m.SummaryPayload, m.SummaryJSON, nil
	}
	return &ChapterSummaryPayload{
		PontosPrincipais: []string{"p1"},
		ResumoGeral:      "r",
		Conceitos:        []string{"c"},
	}, `{"pontos_principais":["p1"],"resumo_geral":"r","conceitos_importantes":["c"]}`, nil
}

func (m *MockService) GenerateFlashcards(_ context.Context, _ string, _ int) ([]FlashcardPair, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if len(m.Flashcards) > 0 {
		return m.Flashcards, nil
	}
	return []FlashcardPair{{Question: "Q?", Answer: "A."}}, nil
}
