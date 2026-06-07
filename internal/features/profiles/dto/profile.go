package dto

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type DoseScheduleDTO struct {
	ID        uuid.UUID `json:"id"`
	ProfileID uuid.UUID `json:"profile_id"`
	Name      string    `json:"name"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProfileDTO struct {
	ID          uuid.UUID         `json:"id"`
	OwnerUserID uuid.UUID         `json:"owner_user_id"`
	Name        string            `json:"name"`
	DateOfBirth *string           `json:"date_of_birth"`
	Timezone    string            `json:"timezone"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Schedules   []DoseScheduleDTO `json:"schedules"`
}

type CreateProfileInput struct {
	Authorization string `header:"Authorization" json:"-"`
	Body          struct {
		Name          string              `json:"name" minLength:"1" maxLength:"100"`
		DateOfBirth   *string             `json:"date_of_birth" format:"date"`
		Timezone      string              `json:"timezone" minLength:"1" maxLength:"50"`
		DoseSchedules []DoseScheduleInput `json:"dose_schedules"`
	}
}

type CreateProfileOutput struct {
	Body struct {
		Profile ProfileDTO `json:"profile"`
	}
}

type ListProfilesOutput struct {
	Body struct {
		Profiles []ProfileDTO `json:"profiles"`
	}
}

type ListProfilesInput struct {
	Authorization string `header:"Authorization" json:"-"`
}

type GetProfileInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
}

type GetProfileOutput struct {
	Body struct {
		Profile ProfileDTO `json:"profile"`
	}
}

type UpdateProfileInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
	Body          struct {
		Name        *string `json:"name" minLength:"1" maxLength:"100"`
		DateOfBirth *string `json:"date_of_birth" format:"date"`
		Timezone    *string `json:"timezone" minLength:"1" maxLength:"50"`
	}
}

type UpdateProfileOutput struct {
	Body struct {
		Profile ProfileDTO `json:"profile"`
	}
}

type DeleteProfileInput struct {
	Authorization string `header:"Authorization" json:"-"`
	ID            string `path:"id"`
}

type DeleteProfileOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type DoseScheduleInput struct {
	Name string `json:"name" minLength:"1" maxLength:"100"`
	Time string `json:"time" format:"time"`
}

func DateOfBirthToPtr(t sql.NullTime) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format("2006-01-02")
	return &s
}
