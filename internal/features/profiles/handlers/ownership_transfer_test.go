package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/handlers"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInitiateTransferHandler_Success(t *testing.T) {
	mockSvc := new(MockOwnershipTransferService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	toUserID := uuid.UUID{2}
	transferID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("InitiateTransfer", mock.Anything, profileID, userID, toUserID).
		Return(&service.OwnershipTransferResult{
			Transfer: service.OwnershipTransferDTO{
				ID:          transferID,
				ProfileID:   profileID,
				ProfileName: "Test Profile",
				FromUserID:  userID,
				ToUserID:    toUserID,
				Status:      "pending",
				ExpiresAt:   now.AddDate(0, 0, 7),
				CreatedAt:   now,
			},
		}, nil)

	handler := handlers.InitiateTransferHandler(mockSvc, mockTokenSvc)

	input := &dto.InitiateTransferInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.ToUserID = toUserID

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pending", resp.Body.Transfer.Status)
}

func TestInitiateTransferHandler_InvalidToken(t *testing.T) {
	mockSvc := new(MockOwnershipTransferService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "invalid-token").Return(nil, assert.AnError)

	handler := handlers.InitiateTransferHandler(mockSvc, mockTokenSvc)

	input := &dto.InitiateTransferInput{}
	input.Authorization = "Bearer invalid-token"
	input.ID = uuid.New().String()
	input.Body.ToUserID = uuid.UUID{2}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestInitiateTransferHandler_InvalidProfileID(t *testing.T) {
	mockSvc := new(MockOwnershipTransferService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": uuid.UUID{1}.String(),
	}, nil)

	handler := handlers.InitiateTransferHandler(mockSvc, mockTokenSvc)

	input := &dto.InitiateTransferInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = "invalid-uuid"
	input.Body.ToUserID = uuid.UUID{2}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListTransfersHandler_Success(t *testing.T) {
	mockSvc := new(MockOwnershipTransferService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	transferID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ListPendingTransfers", mock.Anything, userID).
		Return([]service.OwnershipTransferResult{
			{
				Transfer: service.OwnershipTransferDTO{
					ID:          transferID,
					ProfileID:   profileID,
					ProfileName: "Test Profile",
					FromUserID:  uuid.UUID{2},
					ToUserID:    userID,
					Status:      "pending",
					ExpiresAt:   now.AddDate(0, 0, 7),
					CreatedAt:   now,
				},
			},
		}, nil)

	handler := handlers.ListTransfersHandler(mockSvc, mockTokenSvc)

	input := &dto.ListTransfersInput{}
	input.Authorization = "Bearer valid-token"

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Body.Transfers, 1)
	assert.Equal(t, "pending", resp.Body.Transfers[0].Status)
}

func TestAcceptTransferHandler_Success(t *testing.T) {
	mockSvc := new(MockOwnershipTransferService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	transferID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("AcceptTransfer", mock.Anything, transferID, userID).Return(nil)

	handler := handlers.AcceptTransferHandler(mockSvc, mockTokenSvc)

	input := &dto.TransferActionInput{}
	input.Authorization = "Bearer valid-token"
	input.TransferID = transferID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Transfer accepted successfully", resp.Body.Message)
}

func TestDeclineTransferHandler_Success(t *testing.T) {
	mockSvc := new(MockOwnershipTransferService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	transferID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("DeclineTransfer", mock.Anything, transferID, userID).Return(nil)

	handler := handlers.DeclineTransferHandler(mockSvc, mockTokenSvc)

	input := &dto.TransferActionInput{}
	input.Authorization = "Bearer valid-token"
	input.TransferID = transferID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Transfer declined successfully", resp.Body.Message)
}

func TestCancelTransferHandler_Success(t *testing.T) {
	mockSvc := new(MockOwnershipTransferService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	transferID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("CancelTransfer", mock.Anything, transferID, userID).Return(nil)

	handler := handlers.CancelTransferHandler(mockSvc, mockTokenSvc)

	input := &dto.TransferActionInput{}
	input.Authorization = "Bearer valid-token"
	input.TransferID = transferID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Transfer cancelled successfully", resp.Body.Message)
}
