package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Book struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Author     *string `json:"author,omitempty"`
	FilePath   string  `json:"file_path"`
	TotalPages int     `json:"total_pages"`
	UserID     *string `json:"user_id,omitempty"`
}

type BookRepository struct {
	DB *pgxpool.Pool
}

func NewBookRepository(db *pgxpool.Pool) *BookRepository {
	return &BookRepository{DB: db}
}

func (r *BookRepository) SaveBook(ctx context.Context, userID *string, title, author, path string, totalPages int) (string, error) {
	id := uuid.New().String()
	_, err := r.DB.Exec(ctx,
		`INSERT INTO books (id, user_id, title, author, file_path, total_pages) VALUES ($1,$2,$3,$4,$5,$6)`,
		id, userID, title, nullIfEmpty(author), path, totalPages)
	return id, err
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *BookRepository) SavePage(ctx context.Context, bookID string, page int, content string) error {
	id := uuid.New().String()
	_, err := r.DB.Exec(ctx,
		`INSERT INTO book_pages (id, book_id, page_number, content) VALUES ($1,$2,$3,$4)`,
		id, bookID, page, content)
	return err
}

func (r *BookRepository) GetPageContent(ctx context.Context, bookID string, page int) (string, error) {
	var content string
	err := r.DB.QueryRow(ctx,
		`SELECT content FROM book_pages WHERE book_id=$1 AND page_number=$2`, bookID, page).Scan(&content)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errors.New("página não encontrada")
	}
	return content, err
}

// ConcatPageRange joins logical pages [pageFrom, pageTo] inclusive, in order.
func (r *BookRepository) ConcatPageRange(ctx context.Context, bookID string, pageFrom, pageTo int) (string, error) {
	if pageFrom > pageTo {
		pageFrom, pageTo = pageTo, pageFrom
	}
	rows, err := r.DB.Query(ctx,
		`SELECT content FROM book_pages WHERE book_id=$1 AND page_number >= $2 AND page_number <= $3 ORDER BY page_number ASC`,
		bookID, pageFrom, pageTo)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var b strings.Builder
	for rows.Next() {
		var chunk string
		if err := rows.Scan(&chunk); err != nil {
			return "", err
		}
		b.WriteString(chunk)
		b.WriteString("\n\n")
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if b.Len() == 0 {
		return "", errors.New("nenhuma página no intervalo")
	}
	return strings.TrimSpace(b.String()), nil
}

func (r *BookRepository) ListBooks(ctx context.Context, userID string) ([]Book, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, title, COALESCE(author,''), file_path, total_pages, user_id FROM books WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Book
	for rows.Next() {
		var b Book
		var author string
		if err := rows.Scan(&b.ID, &b.Title, &author, &b.FilePath, &b.TotalPages, &b.UserID); err != nil {
			return nil, err
		}
		if author != "" {
			b.Author = &author
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BookRepository) GetBook(ctx context.Context, bookID string) (*Book, error) {
	var b Book
	var author string
	err := r.DB.QueryRow(ctx,
		`SELECT id, title, COALESCE(author,''), file_path, total_pages, user_id FROM books WHERE id=$1`, bookID).Scan(
		&b.ID, &b.Title, &author, &b.FilePath, &b.TotalPages, &b.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("livro não encontrado")
		}
		return nil, err
	}
	if author != "" {
		b.Author = &author
	}
	return &b, nil
}

// CountByUserID conta livros associados ao usuário.
func (r *BookRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM books WHERE user_id=$1`, userID).Scan(&n)
	return n, err
}

// BelongsToUser indica se o livro pertence ao usuário (user_id igual).
func (r *BookRepository) BelongsToUser(ctx context.Context, bookID, userID string) (bool, error) {
	var uid *string
	err := r.DB.QueryRow(ctx, `SELECT user_id FROM books WHERE id=$1`, bookID).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if uid == nil {
		return false, nil
	}
	return *uid == userID, nil
}
