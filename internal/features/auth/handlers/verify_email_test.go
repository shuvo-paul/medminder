package handlers_test

import (
	"context"
	"testing"

	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVerifyEmail_Successful(t *testing.T) {
	mockSvc := new(MockEmailVerificationService)

	accessToken := "new-access-token"

	mockSvc.On("VerifyEmail", mock.Anything, "valid-token").Return(&service.VerifyResult{
		AccessToken: accessToken,
	}, nil)

	handler := handlers.VerifyEmailHandler(mockSvc)

	input := &dto.VerifyEmailInput{}
	input.Body.Token = "valid-token"

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accessToken, resp.Body.AccessToken)
	mockSvc.AssertExpectations(t)
}

func TestVerifyEmail_EmptyToken(t *testing.T) {
	mockSvc := new(MockEmailVerificationService)
	handler := handlers.VerifyEmailHandler(mockSvc)

	input := &dto.VerifyEmailInput{}
	input.Body.Token = ""

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Token is required")
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	mockSvc := new(MockEmailVerificationService)

	mockSvc.On("VerifyEmail", mock.Anything, "invalid-token").Return(nil, repository.ErrTokenNotFound)

	handler := handlers.VerifyEmailHandler(mockSvc)

	input := &dto.VerifyEmailInput{}
	input.Body.Token = "invalid-token"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockSvc.AssertExpectations(t)
}

func TestVerifyEmail_ExpiredToken(t *testing.T) {
	mockSvc := new(MockEmailVerificationService)

	mockSvc.On("VerifyEmail", mock.Anything, "expired-token").Return(nil, repository.ErrTokenExpired)

	handler := handlers.VerifyEmailHandler(mockSvc)

	input := &dto.VerifyEmailInput{}
	input.Body.Token = "expired-token"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockSvc.AssertExpectations(t)
}
