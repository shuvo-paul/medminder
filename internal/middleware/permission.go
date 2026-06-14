package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	ctxKeyUserID ctxKey = "user_id"
)

var errUnauthorized = errors.New("unauthorized")

type PermissionChecker interface {
	HasAnyPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permissions []string) (bool, error)
}

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (jwt.MapClaims, error)
}

func UserIDFromContext(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ctxKeyUserID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

func HumaRequireProfilePermission(checker PermissionChecker, tokenValidator TokenValidator, permissions ...string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		profileIDStr := ctx.Param("id")
		if profileIDStr == "" {
			next(ctx)
			return
		}

		profileID, err := uuid.Parse(profileIDStr)
		if err != nil {
			writeHumaError(ctx, http.StatusBadRequest, "Invalid profile ID")
			return
		}

		authHeader := ctx.Header("Authorization")
		userID, err := extractUserIDFromHeader(authHeader, tokenValidator)
		if err != nil {
			writeHumaError(ctx, http.StatusUnauthorized, "Invalid or expired access token")
			return
		}

		allowed, err := checker.HasAnyPermission(ctx.Context(), profileID, userID, permissions)
		if err != nil {
			writeHumaError(ctx, http.StatusInternalServerError, "Failed to check permissions")
			return
		}

		if !allowed {
			writeHumaError(ctx, http.StatusForbidden, "Insufficient permissions to access this profile")
			return
		}

		ctx = huma.WithValue(ctx, ctxKeyUserID, userID)
		next(ctx)
	}
}

func extractUserIDFromHeader(authHeader string, validator TokenValidator) (uuid.UUID, error) {
	if len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
		return uuid.Nil, errUnauthorized
	}

	tokenString := authHeader[7:]
	claims, err := validator.ValidateAccessToken(tokenString)
	if err != nil {
		return uuid.Nil, errUnauthorized
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errUnauthorized
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil || userID == uuid.Nil {
		return uuid.Nil, errUnauthorized
	}

	return userID, nil
}

func writeHumaError(ctx huma.Context, status int, detail string) {
	ctx.SetStatus(status)
	ctx.SetHeader("Content-Type", "application/problem+json")
	json.NewEncoder(ctx.BodyWriter()).Encode(map[string]interface{}{
		"title":  http.StatusText(status),
		"status": status,
		"detail": detail,
	})
}
