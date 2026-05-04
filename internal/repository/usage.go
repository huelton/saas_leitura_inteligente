package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsageRepository struct {
	DB *pgxpool.Pool
}

func NewUsageRepository(db *pgxpool.Pool) *UsageRepository {
	return &UsageRepository{DB: db}
}

// DayUTC retorna a data UTC (sem hora) para contagem diária.
func DayUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

func (r *UsageRepository) AddQuestions(ctx context.Context, userID string, dayUTC time.Time, n int) error {
	if n <= 0 {
		return nil
	}
	d := DayUTC(dayUTC)
	_, err := r.DB.Exec(ctx,
		`INSERT INTO usage_daily (user_id, day_utc, questions_generated) VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, day_utc) DO UPDATE SET
		   questions_generated = usage_daily.questions_generated + EXCLUDED.questions_generated`,
		userID, d, n)
	return err
}

func (r *UsageRepository) QuestionsToday(ctx context.Context, userID string) (int, error) {
	d := DayUTC(time.Now().UTC())
	var n int
	err := r.DB.QueryRow(ctx,
		`SELECT COALESCE(questions_generated,0) FROM usage_daily WHERE user_id=$1 AND day_utc=$2`,
		userID, d).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return n, err
}
