package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "ErrNotFound returns 404",
			err:        ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrUnauthorized returns 401",
			err:        ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "ErrForbidden returns 403",
			err:        ErrForbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "ErrConflict returns 409",
			err:        ErrConflict,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "ErrBadRequest returns 400",
			err:        ErrBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrInternal returns 500",
			err:        ErrInternal,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "unknown error returns 500",
			err:        errors.New("unknown"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrapped ErrNotFound returns 404",
			err:        fmt.Errorf("getting user: %w", ErrNotFound),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wrapped ErrBadRequest returns 400",
			err:        fmt.Errorf("validating: %w", ErrBadRequest),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "double wrapped ErrConflict returns 409",
			err:        fmt.Errorf("service: %w", fmt.Errorf("repo: %w", ErrConflict)),
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusCode(tt.err)
			assert.Equal(t, tt.wantStatus, got)
		})
	}
}
