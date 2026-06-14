package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateGuestAccessInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
	Body          struct {
		Label         *string  `json:"label,omitempty" maxLength:"100"`
		Permissions   []string `json:"permissions,omitempty"`
		ExpiresInDays *int     `json:"expires_in_days,omitempty" minimum:"1" maximum:"365"`
	}
}

type CreateGuestAccessOutput struct {
	Body struct {
		ID        uuid.UUID `json:"id"`
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
		Label     string    `json:"label,omitempty"`
	}
}

type GuestAccessTokenDTO struct {
	ID         uuid.UUID  `json:"id"`
	Label      string     `json:"label,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type ListGuestAccessTokensInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
}

type ListGuestAccessTokensOutput struct {
	Body struct {
		Tokens []GuestAccessTokenDTO `json:"tokens"`
	}
}

type RevokeGuestAccessInput struct {
	Authorization string `header:"Authorization" json:"-"`
	TokenID       string `path:"tokenId"`
}

type RevokeGuestAccessOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type GuestMedicationInput struct {
	Token string `path:"token"`
}

type GuestMedicationDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GuestMedicationsOutput struct {
	Body struct {
		Medications []GuestMedicationDTO `json:"medications"`
	}
}

type GuestReminderInput struct {
	Token      string `path:"token"`
	ReminderID string `path:"reminderId"`
}

type GuestReminderDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GuestReminderOutput struct {
	Body struct {
		Reminder GuestReminderDTO `json:"reminder"`
	}
}
