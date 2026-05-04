package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/middleware"
	"github.com/huelton/leitura-inteligente/internal/repository"
)

func (s *Server) requireUserID(c *gin.Context) (string, bool) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return "", false
	}
	return uid, true
}

func (s *Server) assertBookOwner(c *gin.Context, bookID string) bool {
	uid, ok := s.requireUserID(c)
	if !ok {
		return false
	}
	br := repository.NewBookRepository(s.DB)
	belongs, err := br.BelongsToUser(c.Request.Context(), bookID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	if !belongs {
		c.JSON(http.StatusForbidden, gin.H{"error": "acesso negado a este livro"})
		return false
	}
	return true
}
