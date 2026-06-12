package dto

import (
	"time"

	"github.com/google/uuid"
)

type InvitationDTO struct {
	ID               uuid.UUID  `json:"id"`
	ProfileID        uuid.UUID  `json:"profile_id"`
	ProfileName      string     `json:"profile_name"`
	SharedWithUserID uuid.UUID  `json:"shared_with_user_id"`
	GrantedByUserID  uuid.UUID  `json:"granted_by_user_id"`
	Permissions      []string   `json:"permissions"`
	Status           string     `json:"status"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type ShareProfileInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
	Body          struct {
		SharedWithUserID uuid.UUID `json:"shared_with_user_id"`
		Permissions      []string  `json:"permissions" minLength:"1"`
		ExpiresInDays    int       `json:"expires_in_days" minimum:"1"`
	}
}

type ShareProfileOutput struct {
	Body struct {
		Invitation InvitationDTO `json:"invitation"`
	}
}

type ListInvitationsInput struct {
	Authorization string `header:"Authorization" json:"-"`
}

type ListInvitationsOutput struct {
	Body struct {
		Invitations []InvitationDTO `json:"invitations"`
	}
}

type AcceptInvitationInput struct {
	Authorization string `header:"Authorization" json:"-"`
	InvitationID  string `path:"invitationId"`
}

type AcceptInvitationOutput struct {
	Body struct {
		Profile     ProfileDTO `json:"profile"`
		Permissions []string   `json:"permissions"`
	}
}

type DeclineInvitationInput struct {
	Authorization string `header:"Authorization" json:"-"`
	InvitationID  string `path:"invitationId"`
}

type DeclineInvitationOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}
