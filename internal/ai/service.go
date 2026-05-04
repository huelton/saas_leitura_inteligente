package ai

import "context"

// Service abstrai chamadas à IA (implementação real ou mock em testes).
type Service interface {
	GenerateQuestions(ctx context.Context, text string) ([]GeneratedQuestion, error)
	GenerateQuestionsMCQ(ctx context.Context, text string, count int) ([]MCQQuestion, error)
	EvaluateAnswer(ctx context.Context, question, answer string) (int, string, error)
	GenerateChapterSummary(ctx context.Context, text string) (*ChapterSummaryPayload, string, error)
	GenerateFlashcards(ctx context.Context, text string, max int) ([]FlashcardPair, error)
}

var _ Service = (*Client)(nil)
