package handler

import (
	"context"

	"github.com/huelton/leitura-inteligente/internal/ai"
	"github.com/huelton/leitura-inteligente/internal/jobs"
	"github.com/huelton/leitura-inteligente/internal/limiter"
	"github.com/huelton/leitura-inteligente/internal/storage"
	"github.com/huelton/leitura-inteligente/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server agrupa dependências HTTP.
type Server struct {
	DB      *pgxpool.Pool
	Config  *config.Config
	AI      ai.Service
	Jobs    *jobs.Worker
	Limits  limiter.Config
	Storage storage.ObjectStore

	mcqPrewarmQueue chan mcqPrewarmTask
}

func NewServer(db *pgxpool.Pool, cfg *config.Config, aiSvc ai.Service, worker *jobs.Worker, store storage.ObjectStore) *Server {
	s := &Server{
		DB:      db,
		Config:  cfg,
		AI:      aiSvc,
		Jobs:    worker,
		Limits:  cfg.Limits,
		Storage: store,
		// Fila dedicada ao pré-aquecimento de MCQ por leitura de página.
		mcqPrewarmQueue: make(chan mcqPrewarmTask, 128),
	}
	workers := 2
	for i := 0; i < workers; i++ {
		go s.runMCQPrewarmWorker(context.Background())
	}
	return s
}
