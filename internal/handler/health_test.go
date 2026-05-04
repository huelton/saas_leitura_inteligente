package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := &Server{}
	r.GET("/health", s.Health)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("status %d", rw.Code)
	}
}
