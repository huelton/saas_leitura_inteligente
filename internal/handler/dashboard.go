package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/repository"
	"github.com/jackc/pgx/v5"
)

func (s *Server) Dashboard(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	bookID := c.Param("bookId")
	if !s.assertBookOwner(c, bookID) {
		return
	}
	qRepo := repository.NewQuestionRepository(s.DB)

	var score float64
	err := s.DB.QueryRow(c.Request.Context(),
		`SELECT score FROM reading_scores WHERE user_id=$1 AND book_id=$2`, uid, bookID).Scan(&score)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			score = 0
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	pagesRead, _ := qRepo.CountDistinctPagesRead(c.Request.Context(), uid, bookID)
	answers, _ := qRepo.CountAnswersForBook(c.Request.Context(), uid, bookID)

	c.JSON(http.StatusOK, gin.H{
		"user_id":             uid,
		"book_id":             bookID,
		"comprehension_score": score,
		"pages_read_distinct": pagesRead,
		"answers_submitted":   answers,
	})
}
