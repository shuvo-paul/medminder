// Package repository provides data access for audit log entries.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	db "github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/sqlc-dev/pqtype"
)

// AuditRepository defines the interface for audit log data access.
type AuditRepository interface {
	LogEvent(ctx context.Context, eventType string, userID uuid.NullUUID, ipAddress, userAgent string, metadata map[string]string) error
}

// auditRepository implements AuditRepository.
type auditRepository struct {
	queries *db.Queries
}

// NewAuditRepository creates a new audit log repository.
func NewAuditRepository(queries *db.Queries) AuditRepository {
	return &auditRepository{queries: queries}
}

func (r *auditRepository) LogEvent(ctx context.Context, eventType string, userID uuid.NullUUID, ipAddress, userAgent string, metadata map[string]string) error {
	var metaRaw pqtype.NullRawMessage
	if len(metadata) > 0 {
		data, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metaRaw = pqtype.NullRawMessage{RawMessage: data, Valid: true}
	}

	return r.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		ID:        uuid.New(),
		EventType: eventType,
		UserID:    userID,
		IpAddress: sql.NullString{String: ipAddress, Valid: ipAddress != ""},
		UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
		Metadata:  metaRaw,
	})
}
