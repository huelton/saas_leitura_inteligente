package limiter

import (
	"context"
	"errors"
	"fmt"

	"github.com/huelton/leitura-inteligente/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config limites do plano free (0 = ilimitado).
type Config struct {
	FreeMaxBooks        int
	FreeMaxQuestionsDay int
}

// ExceedsFreeBookLimit indica se o usuário free já atingiu o teto de livros (regra pura, fácil de testar).
func ExceedsFreeBookLimit(plan string, cfg Config, bookCount int) bool {
	if plan != "free" || cfg.FreeMaxBooks <= 0 {
		return false
	}
	return bookCount >= cfg.FreeMaxBooks
}

// ExceedsFreeQuestionLimit indica se novas perguntas ultrapassam o teto diário free.
func ExceedsFreeQuestionLimit(plan string, cfg Config, usedToday, newQuestions int) bool {
	if plan != "free" || cfg.FreeMaxQuestionsDay <= 0 {
		return false
	}
	return usedToday+newQuestions > cfg.FreeMaxQuestionsDay
}

// CheckBookUpload retorna erro se o plano free exceder o máximo de livros.
func CheckBookUpload(ctx context.Context, db *pgxpool.Pool, cfg Config, userID, plan string) error {
	if plan != "free" || cfg.FreeMaxBooks <= 0 {
		return nil
	}
	br := repository.NewBookRepository(db)
	n, err := br.CountByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if ExceedsFreeBookLimit(plan, cfg, n) {
		return errors.New("limite de livros do plano free atingido; faça upgrade")
	}
	return nil
}

// CheckQuestions retorna erro se exceder perguntas/dia no free.
func CheckQuestions(ctx context.Context, db *pgxpool.Pool, cfg Config, plan, userID string, newQuestions int) error {
	if plan != "free" || cfg.FreeMaxQuestionsDay <= 0 {
		return nil
	}
	ur := repository.NewUsageRepository(db)
	used, err := ur.QuestionsToday(ctx, userID)
	if err != nil {
		return err
	}
	if ExceedsFreeQuestionLimit(plan, cfg, used, newQuestions) {
		return fmt.Errorf("limite diário de %d perguntas (free) atingido", cfg.FreeMaxQuestionsDay)
	}
	return nil
}
