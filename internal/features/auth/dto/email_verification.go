package dto

// VerifyEmailRequest is the input for verifying an email address.
type VerifyEmailRequest struct {
	Body struct {
		Token string `json:"token" minLength:"1" maxLength:"64"`
	}
}

// VerifyEmailResponse is the output after successful email verification.
type VerifyEmailResponse struct {
	Body struct {
		AccessToken string `json:"access_token"`
	}
}

// ResendVerificationRequest is the input for resending a verification email.
type ResendVerificationRequest struct {
	Body struct {
		Email string `json:"email" minLength:"1" maxLength:"255" pattern:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
	}
}

// ResendVerificationResponse is the output after requesting a verification email resend.
type ResendVerificationResponse struct {
	Body struct {
		Message string `json:"message"`
	}
}