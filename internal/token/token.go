package token

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bekasyldev/lumea/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

var jwtAccessSecret = []byte(os.Getenv("JWT_ACCESS_SECRET"))
var jwtRefreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))

func IssueAccessToken(username string, role string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role: role,
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtAccessSecret)
}

func IssueRefreshToken(username string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add((7 * 24 * time.Hour))),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.NewString(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtRefreshSecret)
}

func ValidateToken(tokenString string, tokenType TokenType, claims *Claims) (*jwt.Token, error) {
	secret := jwtAccessSecret
	if tokenType == TokenRefresh {
		secret = jwtRefreshSecret
	}
	return jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
}

func Persist(ctx context.Context, r *store.Redis, t Claims) error {
	if err := r.SetJTI(ctx, "refresh:"+t.ID, t.Subject, t.ExpiresAt.Time); err != nil {
		return err
	}
	return nil
}
