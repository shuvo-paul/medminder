package service_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/stretchr/testify/assert"
)

func TestPermissionChecker_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		mockFunc   func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error)
		permission string
		expected   bool
	}{
		{
			name: "user_has_permission",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:read", "profile:write"})
				return db.ProfilePermission{
					Status:      "accepted",
					Permissions: perms,
				}, nil
			},
			permission: "profile:read",
			expected:   true,
		},
		{
			name: "user_missing_permission",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:read"})
				return db.ProfilePermission{
					Status:      "accepted",
					Permissions: perms,
				}, nil
			},
			permission: "profile:admin",
			expected:   false,
		},
		{
			name: "no_permission_row",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				return db.ProfilePermission{}, sql.ErrNoRows
			},
			permission: "profile:read",
			expected:   false,
		},
		{
			name: "status_is_pending",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:read"})
				return db.ProfilePermission{
					Status:      "pending",
					Permissions: perms,
				}, nil
			},
			permission: "profile:read",
			expected:   false,
		},
		{
			name: "status_is_declined",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:read"})
				return db.ProfilePermission{
					Status:      "declined",
					Permissions: perms,
				}, nil
			},
			permission: "profile:read",
			expected:   false,
		},
		{
			name: "permission_expired",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:read"})
				return db.ProfilePermission{
					Status:      "accepted",
					Permissions: perms,
					ExpiresAt:   sql.NullTime{Time: time.Now().Add(-1 * time.Hour), Valid: true},
				}, nil
			},
			permission: "profile:read",
			expected:   false,
		},
		{
			name: "owner_has_permission",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:owner"})
				return db.ProfilePermission{
					Status:      "accepted",
					Permissions: perms,
				}, nil
			},
			permission: "profile:owner",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := service.NewPermissionCheckerWithFunc(tt.mockFunc)

			profileID := uuid.New()
			userID := uuid.New()
			result, err := checker.HasPermission(context.Background(), profileID, userID, tt.permission)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermissionChecker_HasAnyPermission(t *testing.T) {
	tests := []struct {
		name        string
		mockFunc    func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error)
		permissions []string
		expected    bool
	}{
		{
			name: "user_has_one_of_permissions",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:read", "profile:write"})
				return db.ProfilePermission{
					Status:      "accepted",
					Permissions: perms,
				}, nil
			},
			permissions: []string{"profile:read", "profile:admin"},
			expected:    true,
		},
		{
			name: "user_has_none_of_permissions",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:read"})
				return db.ProfilePermission{
					Status:      "accepted",
					Permissions: perms,
				}, nil
			},
			permissions: []string{"profile:admin", "profile:write"},
			expected:    false,
		},
		{
			name: "empty_permission_list",
			mockFunc: func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error) {
				perms, _ := json.Marshal([]string{"profile:read"})
				return db.ProfilePermission{
					Status:      "accepted",
					Permissions: perms,
				}, nil
			},
			permissions: []string{},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := service.NewPermissionCheckerWithFunc(tt.mockFunc)

			profileID := uuid.New()
			userID := uuid.New()
			result, err := checker.HasAnyPermission(context.Background(), profileID, userID, tt.permissions)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
