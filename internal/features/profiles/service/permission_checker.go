package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type GetProfilePermissionFunc func(ctx context.Context, arg db.GetProfilePermissionParams) (db.ProfilePermission, error)

type PermissionChecker interface {
	HasPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permission string) (bool, error)
	HasAnyPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permissions []string) (bool, error)
}

type permissionChecker struct {
	getProfilePermission GetProfilePermissionFunc
}

func NewPermissionChecker(queries *db.Queries) PermissionChecker {
	return &permissionChecker{
		getProfilePermission: queries.GetProfilePermission,
	}
}

func NewPermissionCheckerWithFunc(getProfilePermission GetProfilePermissionFunc) PermissionChecker {
	return &permissionChecker{
		getProfilePermission: getProfilePermission,
	}
}

func (c *permissionChecker) HasPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permission string) (bool, error) {
	pp, err := c.getProfilePermission(ctx, db.GetProfilePermissionParams{
		ProfileID:        profileID,
		SharedWithUserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if pp.Status != "accepted" {
		return false, nil
	}

	if pp.ExpiresAt.Valid && pp.ExpiresAt.Time.Before(time.Now()) {
		return false, nil
	}

	var perms []string
	if err := json.Unmarshal(pp.Permissions, &perms); err != nil {
		return false, nil
	}

	for _, p := range perms {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

func (c *permissionChecker) HasAnyPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permissions []string) (bool, error) {
	pp, err := c.getProfilePermission(ctx, db.GetProfilePermissionParams{
		ProfileID:        profileID,
		SharedWithUserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if pp.Status != "accepted" {
		return false, nil
	}

	if pp.ExpiresAt.Valid && pp.ExpiresAt.Time.Before(time.Now()) {
		return false, nil
	}

	var userPerms []string
	if err := json.Unmarshal(pp.Permissions, &userPerms); err != nil {
		return false, nil
	}

	permSet := make(map[string]struct{}, len(userPerms))
	for _, p := range userPerms {
		permSet[p] = struct{}{}
	}

	for _, required := range permissions {
		if _, ok := permSet[required]; ok {
			return true, nil
		}
	}

	return false, nil
}
