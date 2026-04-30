package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/limiter"
	"github.com/huelton/leitura-inteligente/internal/middleware"
	"github.com/huelton/leitura-inteligente/internal/repository"
	"github.com/huelton/leitura-inteligente/internal/service"
)

func (s *Server) UploadBookAsync(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	plan := middleware.CurrentPlan(c)
	if err := limiter.CheckBookUpload(c.Request.Context(), s.DB, s.Limits, uid, plan); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := c.Request.ParseMultipartForm(s.Config.Security.MaxUploadBytes); err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "arquivo excede o tamanho máximo permitido"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo obrigatório (field file)"})
		return
	}
	title := c.PostForm("title")
	if title == "" {
		title = file.Filename
	}
	author := c.PostForm("author")

	filePath, err := s.persistUploadedPDF(c, uid, file)
	if err != nil {
		if errors.Is(err, service.ErrNotPDF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jr := repository.NewJobRepository(s.DB)
	jobID, err := jr.Create(c.Request.Context(), uid, filePath, title, author)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s.Jobs != nil {
		s.Jobs.Enqueue(jobID)
	}

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobID,
		"status": "pending",
	})
}

func (s *Server) GetJob(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	id := c.Param("id")
	jr := repository.NewJobRepository(s.DB)
	j, err := jr.GetByIDForUser(c.Request.Context(), id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	out := gin.H{
		"job_id": j.ID,
		"status": j.Status,
		"title":  j.Title,
	}
	if j.BookID != nil {
		out["book_id"] = *j.BookID
	}
	if j.ErrorMessage != nil && *j.ErrorMessage != "" {
		out["error"] = *j.ErrorMessage
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) ListBooks(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	bookRepo := repository.NewBookRepository(s.DB)
	books, err := bookRepo.ListBooks(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"books": books})
}

func (s *Server) GetPage(c *gin.Context) {
	uid, ok := s.requireUserID(c)
	if !ok {
		return
	}
	bookID := c.Param("id")
	if !s.assertBookOwner(c, bookID) {
		return
	}
	pageStr := c.Param("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "página inválida"})
		return
	}
	bookRepo := repository.NewBookRepository(s.DB)
	content, err := bookRepo.GetPageContent(c.Request.Context(), bookID, page)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	s.enqueueMCQPrewarm(mcqPrewarmTask{
		UserID: uid,
		Plan:   middleware.CurrentPlan(c),
		BookID: bookID,
		Page:   page,
	})
	c.JSON(http.StatusOK, gin.H{"book_id": bookID, "page": page, "content": content})
}

func (s *Server) persistUploadedPDF(c *gin.Context, uid string, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	body, err := service.ReaderAfterPDFValidation(src)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".pdf"
	}
	if s.Storage != nil {
		key := fmt.Sprintf("uploads/%s/%d-%s%s",
			uid, time.Now().UTC().UnixNano(), strings.ReplaceAll(strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename)), " ", "_"), ext)
		objRef, err := s.Storage.PutObject(c.Request.Context(), key, body, file.Size, "application/pdf")
		if err != nil {
			return "", err
		}
		return objRef, nil
	}

	if err := os.MkdirAll("uploads", 0o755); err != nil {
		return "", err
	}
	filePath := filepath.Join("uploads", fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), filepath.Base(file.Filename)))
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(dst, body); err != nil {
		_ = dst.Close()
		_ = os.Remove(filePath)
		return "", err
	}
	if err := dst.Close(); err != nil {
		return "", err
	}
	return filePath, nil
}
