package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QuestionRepository struct {
	DB *pgxpool.Pool
}

type MCQRow struct {
	ID         string
	Question   string
	Difficulty string
	Options    []string
	CorrectIdx int
}

type QuestionMeta struct {
	BookID       string
	Page         int
	Question     string
	QuestionType string
	Options      []string
	CorrectIdx   *int
}

func NewQuestionRepository(db *pgxpool.Pool) *QuestionRepository {
	return &QuestionRepository{DB: db}
}

func (r *QuestionRepository) SaveQuestion(ctx context.Context, bookID string, page int, question, difficulty string) (string, error) {
	id := uuid.New().String()
	_, err := r.DB.Exec(ctx,
		`INSERT INTO questions (id, book_id, page_number, question, difficulty, question_type, created_at) VALUES ($1,$2,$3,$4,$5,'open',$6)`,
		id, bookID, page, question, difficulty, time.Now().UTC())
	return id, err
}

func (r *QuestionRepository) SaveMCQQuestion(ctx context.Context, bookID string, page int, q MCQRow) (string, error) {
	id := uuid.New().String()
	opts, err := json.Marshal(q.Options)
	if err != nil {
		return "", fmt.Errorf("marshal options: %w", err)
	}
	_, err = r.DB.Exec(ctx,
		`INSERT INTO questions
		 (id, book_id, page_number, question, difficulty, question_type, options, correct_idx, created_at)
		 VALUES ($1,$2,$3,$4,$5,'mcq',$6::jsonb,$7,$8)`,
		id, bookID, page, q.Question, q.Difficulty, string(opts), q.CorrectIdx, time.Now().UTC())
	return id, err
}

func (r *QuestionRepository) CountMCQByBookPage(ctx context.Context, bookID string, page int) (int, error) {
	var n int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM questions WHERE book_id=$1 AND page_number=$2 AND question_type='mcq'`,
		bookID, page).Scan(&n)
	return n, err
}

func (r *QuestionRepository) ListRandomMCQByBookPage(ctx context.Context, bookID string, page, limit int) ([]MCQRow, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, question, COALESCE(difficulty, ''), options, COALESCE(correct_idx, 0)
		 FROM questions
		 WHERE book_id=$1 AND page_number=$2 AND question_type='mcq'
		 ORDER BY RANDOM()
		 LIMIT $3`, bookID, page, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]MCQRow, 0, limit)
	for rows.Next() {
		var item MCQRow
		var raw []byte
		if err := rows.Scan(&item.ID, &item.Question, &item.Difficulty, &raw, &item.CorrectIdx); err != nil {
			return nil, err
		}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &item.Options)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *QuestionRepository) ListMCQTextsByBookPage(ctx context.Context, bookID string, page int) ([]string, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT question FROM questions
		 WHERE book_id=$1 AND page_number=$2 AND question_type='mcq'`,
		bookID, page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0, 32)
	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (r *QuestionRepository) GetQuestionText(ctx context.Context, questionID string) (string, error) {
	var q string
	err := r.DB.QueryRow(ctx, `SELECT question FROM questions WHERE id=$1`, questionID).Scan(&q)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errors.New("pergunta não encontrada")
	}
	return q, err
}

func (r *QuestionRepository) GetQuestionMeta(ctx context.Context, questionID string) (QuestionMeta, error) {
	var meta QuestionMeta
	var raw []byte
	err := r.DB.QueryRow(ctx,
		`SELECT book_id, page_number, question, COALESCE(question_type, 'open'), options, correct_idx
		 FROM questions WHERE id=$1`, questionID).
		Scan(&meta.BookID, &meta.Page, &meta.Question, &meta.QuestionType, &raw, &meta.CorrectIdx)
	if errors.Is(err, pgx.ErrNoRows) {
		return QuestionMeta{}, errors.New("pergunta não encontrada")
	}
	if err != nil {
		return QuestionMeta{}, err
	}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &meta.Options)
	}
	if meta.QuestionType == "" {
		meta.QuestionType = "open"
	}
	return meta, nil
}

func (r *QuestionRepository) SaveAnswer(ctx context.Context, questionID, userID, answer string, score int, feedback string) error {
	id := uuid.New().String()
	_, err := r.DB.Exec(ctx,
		`INSERT INTO answers (id, question_id, user_id, answer, score, feedback, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, questionID, userID, answer, score, feedback, time.Now().UTC())
	return err
}

func (r *QuestionRepository) UpsertReadingScore(ctx context.Context, userID, bookID string, score float64) error {
	id := uuid.New().String()
	_, err := r.DB.Exec(ctx,
		`INSERT INTO reading_scores (id, user_id, book_id, score, updated_at) VALUES ($1,$2,$3,$4,NOW())
		 ON CONFLICT (user_id, book_id) DO UPDATE SET score = EXCLUDED.score, updated_at = NOW()`,
		id, userID, bookID, score)
	return err
}

func (r *QuestionRepository) AverageScoreForBook(ctx context.Context, userID, bookID string) (float64, error) {
	var avg *float64
	err := r.DB.QueryRow(ctx,
		`SELECT AVG(a.score)::float8 FROM answers a
		 JOIN questions q ON a.question_id = q.id
		 WHERE a.user_id = $1 AND q.book_id = $2`, userID, bookID).Scan(&avg)
	if err != nil {
		return 0, err
	}
	if avg == nil {
		return 0, nil
	}
	return *avg, nil
}

func (r *QuestionRepository) CountAnswersForBook(ctx context.Context, userID, bookID string) (int, error) {
	var n int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM answers a
		 JOIN questions q ON a.question_id = q.id
		 WHERE a.user_id = $1 AND q.book_id = $2`, userID, bookID).Scan(&n)
	return n, err
}

func (r *QuestionRepository) CountDistinctPagesRead(ctx context.Context, userID, bookID string) (int, error) {
	var n int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(DISTINCT page_number) FROM reading_sessions WHERE user_id=$1 AND book_id=$2`,
		userID, bookID).Scan(&n)
	return n, err
}

func (r *QuestionRepository) AddReadingSession(ctx context.Context, userID, bookID string, page, seconds int) error {
	id := uuid.New().String()
	_, err := r.DB.Exec(ctx,
		`INSERT INTO reading_sessions (id, user_id, book_id, page_number, time_spent_seconds, created_at) VALUES ($1,$2,$3,$4,$5,NOW())`,
		id, userID, bookID, page, seconds)
	return err
}
