package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/ai"
	"github.com/huelton/leitura-inteligente/internal/database"
	"github.com/huelton/leitura-inteligente/internal/handler"
	"github.com/huelton/leitura-inteligente/internal/jobs"
	"github.com/huelton/leitura-inteligente/internal/middleware"
	"github.com/huelton/leitura-inteligente/internal/storage"
	"github.com/huelton/leitura-inteligente/pkg/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		log.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.Migrate(context.Background(), pool); err != nil {
		log.Error("migrate", "err", err)
		os.Exit(1)
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	var objectStore storage.ObjectStore
	if cfg.Storage.Provider == "s3" {
		s3Store, err := storage.NewS3Store(context.Background(), cfg.Storage)
		if err != nil {
			log.Error("storage", "err", err)
			os.Exit(1)
		}
		objectStore = s3Store
	}

	w := jobs.NewWorker(pool, cfg.PageChunkSize, 64, objectStore)
	go w.Run(workerCtx)

	aiClient := ai.NewClient(cfg.AIAPIKey, cfg.AIModel)
	srv := handler.NewServer(pool, cfg, aiClient, w, objectStore)

	r := gin.New()
	r.MaxMultipartMemory = 8 << 20
	r.Use(gin.Recovery())
	r.Use(requestLogger(log))
	r.Use(middleware.CORS(cfg.Security.CORSAllowedOrigins))
	srv.RegisterRoutes(r)

	addr := cfg.HTTPAddr
	if addr == "" {
		addr = ":8080"
	}

	go func() {
		log.Info("listening", "addr", addr)
		if err := r.Run(addr); err != nil {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutdown")
	workerCancel()
	time.Sleep(100 * time.Millisecond)
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"ms", time.Since(start).Milliseconds(),
		)
	}
}
