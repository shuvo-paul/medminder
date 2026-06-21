package auth

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrUnauthorized = errors.New("unauthorized")

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (jwt.MapClaims, error)
}

func ExtractUserID(authHeader string, validator TokenValidator) (uuid.UUID, error) {
	if len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
		return uuid.Nil, ErrUnauthorized
	}

	tokenString := authHeader[7:]
	claims, err := validator.ValidateAccessToken(tokenString)
	if err != nil {
		return uuid.Nil, ErrUnauthorized
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, ErrUnauthorized
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil || userID == uuid.Nil {
		return uuid.Nil, ErrUnauthorized
	}

	return userID, nil
}

func ExtractEmail(authHeader string, validator TokenValidator) (string, error) {
	if len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrUnauthorized
	}

	tokenString := authHeader[7:]
	claims, err := validator.ValidateAccessToken(tokenString)
	if err != nil {
		return "", ErrUnauthorized
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return "", ErrUnauthorized
	}

	return email, nil
}
