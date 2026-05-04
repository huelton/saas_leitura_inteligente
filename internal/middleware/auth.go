package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huelton/leitura-inteligente/internal/auth"
)

const (
	ContextUserID   = "auth_user_id"
	ContextEmail    = "auth_email"
	ContextPlan     = "auth_plan"
	HeaderBearer    = "Authorization"
)

// JWTRequired valida Bearer JWT e preenche o contexto Gin.
func JWTRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader(HeaderBearer)
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization Bearer obrigatório"})
			return
		}
		raw := strings.TrimSpace(h[7:])
		claims, err := auth.ParseAccessToken(secret, raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido ou expirado"})
			return
		}
		c.Set(ContextUserID, auth.UserID(claims))
		c.Set(ContextEmail, claims.Email)
		c.Set(ContextPlan, claims.Plan)
		c.Next()
	}
}

// CurrentUserID retorna o id do usuário autenticado (JWT).
func CurrentUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextUserID)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

// CurrentPlan retorna o plano do token (ex.: free).
func CurrentPlan(c *gin.Context) string {
	p, ok := c.Get(ContextPlan)
	if !ok {
		return "free"
	}
	s, ok := p.(string)
	if !ok || s == "" {
		return "free"
	}
	return s
}
