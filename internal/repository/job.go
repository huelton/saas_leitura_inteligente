package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Job struct {
	ID           string
	UserID       string
	Status       string
	FilePath     string
	Title        string
	Author       *string
	ErrorMessage *string
	BookID       *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type JobRepository struct {
	DB *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{DB: db}
}

func (r *JobRepository) Create(ctx context.Context, userID, filePath, title, author string) (string, error) {
	id := uuid.New().String()
	var auth interface{}
	if author != "" {
		auth = author
	}
	_, err := r.DB.Exec(ctx,
		`INSERT INTO processing_jobs (id, user_id, status, file_path, title, author, created_at, updated_at)
		 VALUES ($1,$2,'pending',$3,$4,$5,NOW(),NOW())`,
		id, userID, filePath, title, auth)
	return id, err
}

func (r *JobRepository) GetByID(ctx context.Context, id string) (*Job, error) {
	var j Job
	var auth *string
	err := r.DB.QueryRow(ctx,
		`SELECT id, user_id, status, file_path, title, author, error_message, book_id, created_at, updated_at
		 FROM processing_jobs WHERE id=$1`, id).Scan(
		&j.ID, &j.UserID, &j.Status, &j.FilePath, &j.Title, &auth, &j.ErrorMessage, &j.BookID, &j.CreatedAt, &j.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("job não encontrado")
	}
	j.Author = auth
	return &j, err
}

func (r *JobRepository) GetByIDForUser(ctx context.Context, id, userID string) (*Job, error) {
	j, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if j.UserID != userID {
		return nil, errors.New("job não encontrado")
	}
	return j, nil
}

func (r *JobRepository) SetProcessing(ctx context.Context, id string) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE processing_jobs SET status='processing', updated_at=NOW() WHERE id=$1 AND status='pending'`, id)
	return err
}

func (r *JobRepository) SetDone(ctx context.Context, id, bookID string) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE processing_jobs SET status='done', book_id=$2, updated_at=NOW() WHERE id=$1`, id, bookID)
	return err
}

func (r *JobRepository) SetFailed(ctx context.Context, id, msg string) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE processing_jobs SET status='failed', error_message=$2, updated_at=NOW() WHERE id=$1`, id, msg)
	return err
}
