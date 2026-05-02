package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChapterSummary struct {
	ID             string    `json:"id"`
	BookID         string    `json:"book_id"`
	ChapterNumber  int       `json:"chapter_number"`
	PageFrom       int       `json:"page_from"`
	PageTo         int       `json:"page_to"`
	SummaryJSON    string    `json:"summary_json"`
	CreatedAt      time.Time `json:"created_at"`
}

type ChapterSummaryRepository struct {
	DB *pgxpool.Pool
}

func NewChapterSummaryRepository(db *pgxpool.Pool) *ChapterSummaryRepository {
	return &ChapterSummaryRepository{DB: db}
}

func (r *ChapterSummaryRepository) Upsert(ctx context.Context, bookID string, chapterNumber, pageFrom, pageTo int, summaryJSON string) error {
	id := uuid.New().String()
	_, err := r.DB.Exec(ctx,
		`INSERT INTO chapter_summaries (id, book_id, chapter_number, page_from, page_to, summary, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,NOW())
		 ON CONFLICT (book_id, chapter_number) DO UPDATE SET
		   page_from = EXCLUDED.page_from,
		   page_to = EXCLUDED.page_to,
		   summary = EXCLUDED.summary,
		   created_at = NOW()`,
		id, bookID, chapterNumber, pageFrom, pageTo, summaryJSON)
	return err
}

func (r *ChapterSummaryRepository) ListByBook(ctx context.Context, bookID string) ([]ChapterSummary, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, book_id, chapter_number, page_from, page_to, summary, created_at
		 FROM chapter_summaries WHERE book_id=$1 ORDER BY chapter_number ASC`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChapterSummary
	for rows.Next() {
		var s ChapterSummary
		if err := rows.Scan(&s.ID, &s.BookID, &s.ChapterNumber, &s.PageFrom, &s.PageTo, &s.SummaryJSON, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
