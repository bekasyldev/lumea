package token

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func IssueAccessToken(username string, role string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role: role,
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func IssueRefreshToken(username string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add((7 * 24 * time.Hour))),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func ValidateToken(tokenString string, claims *Claims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
}
