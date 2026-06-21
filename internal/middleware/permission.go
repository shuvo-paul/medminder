package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/common/auth"
	profileService "github.com/shuvo-paul/medminder/internal/features/profiles/service"
)

const (
	ctxKeyUserID ctxKey = "user_id"
)

func UserIDFromContext(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ctxKeyUserID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

func HumaRequireProfilePermission(checker profileService.PermissionChecker, tokenValidator auth.TokenValidator, permissions ...string) func(huma.Context, func(huma.Context)) {
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
		userID, err := auth.ExtractUserID(authHeader, tokenValidator)
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

func writeHumaError(ctx huma.Context, status int, detail string) {
	ctx.SetStatus(status)
	ctx.SetHeader("Content-Type", "application/problem+json")
	json.NewEncoder(ctx.BodyWriter()).Encode(map[string]interface{}{
		"title":  http.StatusText(status),
		"status": status,
		"detail": detail,
	})
}
