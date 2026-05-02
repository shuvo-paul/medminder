package dto

// VerifyEmailInput is the input for verifying an email address.
type VerifyEmailInput struct {
	Body struct {
		Token string `json:"token" minLength:"1" maxLength:"64"`
	}
}

// VerifyEmailOutput is the output after successful email verification.
type VerifyEmailOutput struct {
	Body struct {
		AccessToken string `json:"access_token"`
	}
}

// ResendVerificationInput is the input for resending a verification email.
type ResendVerificationInput struct {
	Body struct {
		Authorization string `json:"authorization"`
	}
}

// ResendVerificationOutput is the output after requesting a verification email resend.
type ResendVerificationOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}