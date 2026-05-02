package repository

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Flashcard struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	BookID         string    `json:"book_id"`
	PageNumber     *int      `json:"page_number,omitempty"`
	Question       string    `json:"question"`
	Answer         string    `json:"answer"`
	IntervalDays   int       `json:"interval_days"`
	EaseFactor     float64   `json:"ease_factor"`
	NextReviewDate time.Time `json:"next_review_date"`
	CreatedAt      time.Time `json:"created_at"`
}

type FlashcardRepository struct {
	DB *pgxpool.Pool
}

func NewFlashcardRepository(db *pgxpool.Pool) *FlashcardRepository {
	return &FlashcardRepository{DB: db}
}

func (r *FlashcardRepository) Insert(ctx context.Context, userID, bookID string, pageNumber *int, question, answer string) (string, error) {
	id := uuid.New().String()
	next := time.Now().UTC()
	_, err := r.DB.Exec(ctx,
		`INSERT INTO flashcards (id, user_id, book_id, page_number, question, answer, interval_days, ease_factor, next_review_date, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,1,2.5,$7,NOW())`,
		id, userID, bookID, pageNumber, question, answer, next)
	return id, err
}

func (r *FlashcardRepository) GetForUser(ctx context.Context, id, userID string) (*Flashcard, error) {
	var f Flashcard
	err := r.DB.QueryRow(ctx,
		`SELECT id, user_id, book_id, page_number, question, answer, interval_days, ease_factor, next_review_date, created_at
		 FROM flashcards WHERE id=$1 AND user_id=$2`, id, userID).Scan(
		&f.ID, &f.UserID, &f.BookID, &f.PageNumber, &f.Question, &f.Answer,
		&f.IntervalDays, &f.EaseFactor, &f.NextReviewDate, &f.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("flashcard não encontrado")
	}
	return &f, err
}

// ApplyReview updates interval, ease and next review (spaced repetition simplificado).
func (r *FlashcardRepository) ApplyReview(ctx context.Context, id, userID string, correct bool) error {
	f, err := r.GetForUser(ctx, id, userID)
	if err != nil {
		return err
	}
	interval, ease := nextIntervalAndEase(f.IntervalDays, f.EaseFactor, correct)
	next := time.Now().UTC().AddDate(0, 0, interval)
	_, err = r.DB.Exec(ctx,
		`UPDATE flashcards SET interval_days=$1, ease_factor=$2, next_review_date=$3 WHERE id=$4 AND user_id=$5`,
		interval, ease, next, id, userID)
	return err
}

func nextIntervalAndEase(interval int, ease float64, correct bool) (int, float64) {
	if correct {
		n := int(float64(interval) * ease)
		if n < 1 {
			n = 1
		}
		if n > 365 {
			n = 365
		}
		ease = math.Min(3.0, ease+0.05)
		return n, ease
	}
	ease = math.Max(1.3, ease-0.2)
	return 1, ease
}

func (r *FlashcardRepository) ListDue(ctx context.Context, userID string, limit int) ([]Flashcard, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.DB.Query(ctx,
		`SELECT id, user_id, book_id, page_number, question, answer, interval_days, ease_factor, next_review_date, created_at
		 FROM flashcards WHERE user_id=$1 AND next_review_date <= NOW()
		 ORDER BY next_review_date ASC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Flashcard
	for rows.Next() {
		var f Flashcard
		if err := rows.Scan(&f.ID, &f.UserID, &f.BookID, &f.PageNumber, &f.Question, &f.Answer,
			&f.IntervalDays, &f.EaseFactor, &f.NextReviewDate, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
