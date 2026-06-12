package dto

import (
	"time"

	"github.com/google/uuid"
)

type OwnershipTransferDTO struct {
	ID          uuid.UUID `json:"id"`
	ProfileID   uuid.UUID `json:"profile_id"`
	ProfileName string    `json:"profile_name"`
	FromUserID  uuid.UUID `json:"from_user_id"`
	FromName    string    `json:"from_name"`
	ToUserID    uuid.UUID `json:"to_user_id"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type InitiateTransferInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
	Body          struct {
		ToUserID uuid.UUID `json:"to_user_id"`
	}
}

type InitiateTransferOutput struct {
	Body struct {
		Transfer OwnershipTransferDTO `json:"transfer"`
	}
}

type ListTransfersInput struct {
	Authorization string `header:"Authorization" json:"-"`
}

type ListTransfersOutput struct {
	Body struct {
		Transfers []OwnershipTransferDTO `json:"transfers"`
	}
}

type TransferActionInput struct {
	Authorization string `header:"Authorization" json:"-"`
	TransferID    string `path:"transferId"`
}

type TransferActionOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}
