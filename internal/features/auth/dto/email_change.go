package dto

type RequestEmailChangeInput struct {
	Authorization string `header:"Authorization" json:"-"`
	Body          struct {
		NewEmail        string `json:"new_email" minLength:"1" maxLength:"255" format:"email"`
		CurrentPassword string `json:"current_password" minLength:"1"`
	}
}

type RequestEmailChangeOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type CancelEmailChangeInput struct {
	Authorization string `header:"Authorization" json:"-"`
}

type CancelEmailChangeOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

type GetPendingEmailChangeInput struct {
	Authorization string `header:"Authorization" json:"-"`
}

type GetPendingEmailChangeOutput struct {
	Body struct {
		NewEmail  string `json:"new_email"`
		ExpiresAt string `json:"expires_at"`
	}
}
