package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/repository"
)

type readingProgressReq struct {
	BookID  string `json:"book_id" binding:"required"`
	Page    int    `json:"page" binding:"required"`
	Seconds int    `json:"seconds" binding:"required"`
}

func (s *Server) ReadingProgress(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	var req readingProgressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !s.assertBookOwner(c, req.BookID) {
		return
	}
	repo := repository.NewQuestionRepository(s.DB)
	if err := repo.AddReadingSession(c.Request.Context(), uid, req.BookID, req.Page, req.Seconds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
