package dto

type PasswordResetRequestInput struct {
	Body struct {
		Email string `json:"email" minLength:"1" maxLength:"255" pattern:"^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$"`
	}
}

type PasswordResetRequestOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type PasswordResetConfirmInput struct {
	Body struct {
		Token       string `json:"token" minLength:"1"`
		NewPassword string `json:"new_password" minLength:"8"`
	}
}

type PasswordResetConfirmOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}
