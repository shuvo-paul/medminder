package dto

import "github.com/google/uuid"

// User represents a user in auth responses.
type User struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	EmailVerified bool      `json:"email_verified"`
}

type RegisterInput struct {
	Body struct {
		Email       string `json:"email" minLength:"1" maxLength:"255" pattern:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
		DisplayName string `json:"display_name" minLength:"1" maxLength:"100"`
		Password    string `json:"password" minLength:"8"`
	}
}

type RegisterOutput struct {
	Body struct {
		User User `json:"user"`
	}
}

type LoginInput struct {
	Body struct {
		Email    string `json:"email" minLength:"1" maxLength:"255"`
		Password string `json:"password" minLength:"1"`
	}
}

type LoginOutput struct {
	Body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         User   `json:"user"`
	}
}

type LogoutInput struct {
	Authorization string `header:"Authorization" json:"-"`
}

type LogoutOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}
