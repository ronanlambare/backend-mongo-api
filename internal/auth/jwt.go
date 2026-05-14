package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID string    `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// GenerateTokenPair creates both access and refresh tokens for a user.
func GenerateTokenPair(userID, email, role, secret string, accessExpiry, refreshExpiry time.Duration) (*TokenPair, error) {
	access, err := generateToken(userID, email, role, secret, AccessToken, accessExpiry)
	if err != nil {
		return nil, err
	}

	refresh, err := generateToken(userID, email, role, secret, RefreshToken, refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func generateToken(userID, email, role, secret string, tokenType TokenType, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "lambare-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateClaims parses and validates a JWT, returning its claims.
func ValidateClaims(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
