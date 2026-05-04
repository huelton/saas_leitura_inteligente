package jobs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/huelton/leitura-inteligente/internal/repository"
	"github.com/huelton/leitura-inteligente/internal/service"
	"github.com/huelton/leitura-inteligente/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Worker processa jobs de upload de PDF em fila (um worker por instância).
type Worker struct {
	DB      *pgxpool.Pool
	Chunk   int
	Queue   chan string
	Logger  *slog.Logger
	Storage storage.ObjectStore
}

func NewWorker(db *pgxpool.Pool, chunk int, buf int, store storage.ObjectStore) *Worker {
	if buf <= 0 {
		buf = 32
	}
	return &Worker{
		DB:      db,
		Chunk:   chunk,
		Queue:   make(chan string, buf),
		Logger:  slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
		Storage: store,
	}
}

func (w *Worker) Enqueue(jobID string) {
	select {
	case w.Queue <- jobID:
	default:
		w.Logger.Warn("job queue full, blocking", "job_id", jobID)
		w.Queue <- jobID
	}
}

// Run bloqueia até ctx cancelado.
func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case jobID := <-w.Queue:
			w.process(ctx, jobID)
		}
	}
}

func (w *Worker) process(ctx context.Context, jobID string) {
	jr := repository.NewJobRepository(w.DB)
	j, err := jr.GetByID(ctx, jobID)
	if err != nil {
		w.Logger.Error("job load", "job_id", jobID, "err", err)
		return
	}
	if j.Status != "pending" {
		return
	}
	if err := jr.SetProcessing(ctx, jobID); err != nil {
		w.Logger.Error("job processing", "job_id", jobID, "err", err)
		return
	}

	pdfPath, cleanup, err := w.resolvePDFPath(ctx, j.FilePath)
	if err != nil {
		_ = jr.SetFailed(ctx, jobID, err.Error())
		return
	}
	if cleanup != nil {
		defer cleanup()
	}
	text, err := service.ExtractTextFromPDF(pdfPath)
	if err != nil {
		_ = jr.SetFailed(ctx, jobID, err.Error())
		return
	}
	pages := service.SplitTextIntoPages(text, w.Chunk)
	bookRepo := repository.NewBookRepository(w.DB)
	uid := j.UserID
	author := ""
	if j.Author != nil {
		author = *j.Author
	}
	bookID, err := bookRepo.SaveBook(ctx, &uid, j.Title, author, j.FilePath, len(pages))
	if err != nil {
		_ = jr.SetFailed(ctx, jobID, err.Error())
		return
	}
	for i, content := range pages {
		if err := bookRepo.SavePage(ctx, bookID, i+1, content); err != nil {
			_ = jr.SetFailed(ctx, jobID, err.Error())
			return
		}
	}
	if err := jr.SetDone(ctx, jobID, bookID); err != nil {
		w.Logger.Error("job done", "job_id", jobID, "err", err)
	}
	if w.Logger != nil {
		w.Logger.Info("job completed", "job_id", jobID, "book_id", bookID, "pages", len(pages))
	}
}

func (w *Worker) resolvePDFPath(ctx context.Context, filePath string) (string, func(), error) {
	if w.Storage == nil {
		return filePath, nil, nil
	}
	if !service.FileExists(filePath) {
		rc, err := w.Storage.GetObject(ctx, filePath)
		if err != nil {
			return "", nil, err
		}
		tmp, err := os.CreateTemp("", "leitura-*.pdf")
		if err != nil {
			_ = rc.Close()
			return "", nil, err
		}
		if _, err := io.Copy(tmp, rc); err != nil {
			_ = rc.Close()
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
			return "", nil, err
		}
		_ = rc.Close()
		_ = tmp.Close()
		path := tmp.Name()
		return path, func() { _ = os.Remove(path) }, nil
	}
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return filePath, nil, nil
	}
	return abs, nil, nil
}
