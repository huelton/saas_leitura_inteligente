package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/repository"
)

type summaryChapterReq struct {
	BookID        string `json:"book_id" binding:"required"`
	ChapterNumber int    `json:"chapter_number" binding:"required"`
	PageFrom      int    `json:"page_from" binding:"required"`
	PageTo        int    `json:"page_to" binding:"required"`
}

func (s *Server) GenerateChapterSummary(c *gin.Context) {
	var req summaryChapterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !s.assertBookOwner(c, req.BookID) {
		return
	}
	if req.PageFrom > req.PageTo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_from não pode ser maior que page_to"})
		return
	}
	bookRepo := repository.NewBookRepository(s.DB)
	text, err := bookRepo.ConcatPageRange(c.Request.Context(), req.BookID, req.PageFrom, req.PageTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload, summaryJSON, err := s.AI.GenerateChapterSummary(c.Request.Context(), text)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	sumRepo := repository.NewChapterSummaryRepository(s.DB)
	if err := sumRepo.Upsert(c.Request.Context(), req.BookID, req.ChapterNumber, req.PageFrom, req.PageTo, summaryJSON); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"summary":        payload,
		"summary_json":   summaryJSON,
		"book_id":        req.BookID,
		"chapter_number": req.ChapterNumber,
		"page_from":      req.PageFrom,
		"page_to":        req.PageTo,
	})
}

func (s *Server) ListChapterSummaries(c *gin.Context) {
	bookID := c.Query("book_id")
	if bookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query book_id obrigatório"})
		return
	}
	if !s.assertBookOwner(c, bookID) {
		return
	}
	sumRepo := repository.NewChapterSummaryRepository(s.DB)
	list, err := sumRepo.ListByBook(c.Request.Context(), bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"summaries": list})
}

type flashcardsPageReq struct {
	BookID   string `json:"book_id" binding:"required"`
	Page     int    `json:"page" binding:"required"`
	MaxCards int    `json:"max_cards"`
}

func (s *Server) GenerateFlashcardsForPage(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	var req flashcardsPageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !s.assertBookOwner(c, req.BookID) {
		return
	}
	bookRepo := repository.NewBookRepository(s.DB)
	content, err := bookRepo.GetPageContent(c.Request.Context(), req.BookID, req.Page)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	pairs, err := s.AI.GenerateFlashcards(c.Request.Context(), content, req.MaxCards)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	fcRepo := repository.NewFlashcardRepository(s.DB)
	page := req.Page
	out := make([]gin.H, 0, len(pairs))
	for _, p := range pairs {
		if p.Question == "" || p.Answer == "" {
			continue
		}
		id, err := fcRepo.Insert(c.Request.Context(), uid, req.BookID, &page, p.Question, p.Answer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, gin.H{"id": id, "question": p.Question, "answer": p.Answer, "page": req.Page})
	}
	c.JSON(http.StatusOK, gin.H{"flashcards": out})
}

type flashcardReviewReq struct {
	FlashcardID string `json:"flashcard_id" binding:"required"`
	Correct     bool   `json:"correct"`
}

func (s *Server) ReviewFlashcard(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	var req flashcardReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fcRepo := repository.NewFlashcardRepository(s.DB)
	if err := fcRepo.ApplyReview(c.Request.Context(), req.FlashcardID, uid, req.Correct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := fcRepo.GetForUser(c.Request.Context(), req.FlashcardID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"flashcard_id":     f.ID,
		"interval_days":    f.IntervalDays,
		"ease_factor":      f.EaseFactor,
		"next_review_date": f.NextReviewDate,
	})
}

func (s *Server) ListDueFlashcards(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	fcRepo := repository.NewFlashcardRepository(s.DB)
	list, err := fcRepo.ListDue(c.Request.Context(), uid, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"due": list})
}
