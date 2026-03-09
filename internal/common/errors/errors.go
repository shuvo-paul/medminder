package errors

import (
	"errors"
	"net/http"
)

var (
	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = errors.New("not found")
	// ErrUnauthorized is returned when authentication credentials are invalid.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden is returned when the user lacks permission to access a resource.
	ErrForbidden = errors.New("forbidden")
	// ErrConflict is returned when a resource already exists or conflicts with existing state.
	ErrConflict = errors.New("conflict")
	// ErrBadRequest is returned when the request is invalid or malformed.
	ErrBadRequest = errors.New("bad request")
	// ErrInternal is returned for unexpected internal server errors.
	ErrInternal = errors.New("internal error")
)

func StatusCode(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
