package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", s.Health)

	sec := s.Config.Security
	authRL := middleware.RateLimitPerIP(sec.AuthRatePerMin)
	upRL := middleware.RateLimitPerIP(sec.UploadRatePerMin)
	mAuth := middleware.MetricsBasicAuth(sec.MetricsBasicUser, sec.MetricsBasicPass)
	if sec.MetricsEnabled {
		r.GET("/metrics", mAuth, gin.WrapH(promhttp.Handler()))
	}

	r.POST("/auth/register", authRL, s.Register)
	r.POST("/auth/login", authRL, s.Login)

	authz := r.Group("")
	authz.Use(middleware.JWTRequired(s.Config.JWTSecret))

	authz.GET("/books", s.ListBooks)
	authz.POST("/books/upload", upRL, s.UploadBookAsync)
	authz.GET("/jobs/:id", s.GetJob)
	authz.GET("/books/:id/pages/:page", s.GetPage)

	authz.POST("/reading/progress", s.ReadingProgress)

	authz.POST("/ai/questions/page", s.GenerateQuestionsByPage)
	authz.POST("/ai/questions/page-mcq", s.GenerateQuestionsByPageMCQ)
	authz.POST("/answers", s.SubmitAnswer)

	authz.GET("/summaries", s.ListChapterSummaries)
	authz.POST("/ai/summary/chapter", s.GenerateChapterSummary)
	authz.POST("/ai/flashcards/page", s.GenerateFlashcardsForPage)
	authz.POST("/flashcards/review", s.ReviewFlashcard)
	authz.GET("/flashcards/due", s.ListDueFlashcards)

	authz.GET("/dashboard/:bookId", s.Dashboard)
}

func (s *Server) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
