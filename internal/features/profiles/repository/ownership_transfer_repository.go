package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type OwnershipTransferRepository interface {
	CreateTransfer(ctx context.Context, profileID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID, expiresAt time.Time) (db.OwnershipTransfer, error)
	GetTransferByID(ctx context.Context, transferID uuid.UUID) (db.OwnershipTransfer, error)
	GetPendingTransferByProfile(ctx context.Context, profileID uuid.UUID) (db.OwnershipTransfer, error)
	ListPendingTransfersByUser(ctx context.Context, userID uuid.UUID) ([]db.OwnershipTransfer, error)
	UpdateTransferStatus(ctx context.Context, transferID uuid.UUID, status string) (db.OwnershipTransfer, error)
}

type ownershipTransferRepository struct {
	queries *db.Queries
	db      *sql.DB
}

func NewOwnershipTransferRepository(queries *db.Queries, db *sql.DB) OwnershipTransferRepository {
	return &ownershipTransferRepository{queries: queries, db: db}
}

func (r *ownershipTransferRepository) CreateTransfer(ctx context.Context, profileID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID, expiresAt time.Time) (db.OwnershipTransfer, error) {
	return r.queries.CreateOwnershipTransfer(ctx, db.CreateOwnershipTransferParams{
		ProfileID:  profileID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Status:     "pending",
		ExpiresAt:  expiresAt,
	})
}

func (r *ownershipTransferRepository) GetTransferByID(ctx context.Context, transferID uuid.UUID) (db.OwnershipTransfer, error) {
	return r.queries.GetOwnershipTransferByID(ctx, transferID)
}

func (r *ownershipTransferRepository) GetPendingTransferByProfile(ctx context.Context, profileID uuid.UUID) (db.OwnershipTransfer, error) {
	return r.queries.GetPendingTransferByProfile(ctx, profileID)
}

func (r *ownershipTransferRepository) ListPendingTransfersByUser(ctx context.Context, userID uuid.UUID) ([]db.OwnershipTransfer, error) {
	return r.queries.ListPendingTransfersByUser(ctx, userID)
}

func (r *ownershipTransferRepository) UpdateTransferStatus(ctx context.Context, transferID uuid.UUID, status string) (db.OwnershipTransfer, error) {
	return r.queries.UpdateOwnershipTransferStatus(ctx, db.UpdateOwnershipTransferStatusParams{
		ID:     transferID,
		Status: status,
	})
}
