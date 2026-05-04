package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Email string `json:"email"`
	Plan  string `json:"plan"`
	jwt.RegisteredClaims
}

func SignAccessToken(secret string, userID, email, plan string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("JWT_SECRET is empty")
	}
	now := time.Now()
	claims := Claims{
		Email: email,
		Plan:  plan,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func ParseAccessToken(secret, tokenString string) (*Claims, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET is empty")
	}
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}

// UserID extrai o subject (id do usuário) dos claims.
func UserID(c *Claims) string {
	if c == nil {
		return ""
	}
	return c.Subject
}
