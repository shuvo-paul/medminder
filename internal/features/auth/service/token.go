package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// AccessTokenExpiry is the lifetime of access tokens.
	AccessTokenExpiry = 24 * time.Hour
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	refreshTokenByteSize = 32
)

type TokenService struct {
	jwtSecret []byte
}

var _ TokenServiceInterface = (*TokenService)(nil)

type TokenServiceInterface interface {
	GenerateAccessToken(userID uuid.UUID, email string) (string, error)
	GenerateRefreshToken() (string, error)
	HashRefreshToken(token string) string
	ValidateAccessToken(tokenString string) (jwt.MapClaims, error)
}

func NewTokenService(jwtSecret string) *TokenService {
	return &TokenService{jwtSecret: []byte(jwtSecret)}
}

func (ts *TokenService) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"exp":   time.Now().Add(AccessTokenExpiry).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ts.jwtSecret)
}

func (ts *TokenService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, refreshTokenByteSize)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (ts *TokenService) HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (ts *TokenService) VerifyRefreshToken(token, hash string) error {
	if ts.HashRefreshToken(token) != hash {
		return ErrInvalidToken
	}
	return nil
}

func (ts *TokenService) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ts.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > int64(exp) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}
