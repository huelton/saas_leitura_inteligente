package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/auth"
	"github.com/huelton/leitura-inteligente/internal/repository"
)

type registerReq struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s.Config.JWTSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "JWT_SECRET não configurado no servidor"})
		return
	}
	repo := repository.NewUserRepository(s.DB)
	u, err := repo.Create(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	token, err := auth.SignAccessToken(s.Config.JWTSecret, u.ID, u.Email, u.Plan, s.Config.JWTExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   int(s.Config.JWTExpiry.Seconds()),
		"user_id":      u.ID,
		"name":         u.Name,
		"email":        u.Email,
		"plan":         u.Plan,
	})
}

func (s *Server) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s.Config.JWTSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "JWT_SECRET não configurado no servidor"})
		return
	}
	repo := repository.NewUserRepository(s.DB)
	u, err := repo.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, err := auth.SignAccessToken(s.Config.JWTSecret, u.ID, u.Email, u.Plan, s.Config.JWTExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   int(s.Config.JWTExpiry.Seconds()),
		"user_id":      u.ID,
		"name":         u.Name,
		"email":        u.Email,
		"plan":         u.Plan,
	})
}
