package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockEmailChangeRepository mocks repository.EmailChangeRepository
type MockEmailChangeRepository struct {
	mock.Mock
}

func (m *MockEmailChangeRepository) Create(ctx context.Context, userID uuid.UUID, newEmail, tokenHash string, expiresAt time.Time) (db.EmailChangeRequest, error) {
	args := m.Called(ctx, userID, newEmail, tokenHash, expiresAt)
	return args.Get(0).(db.EmailChangeRequest), args.Error(1)
}

func (m *MockEmailChangeRepository) FindValidByTokenHash(ctx context.Context, tokenHash string) (db.EmailChangeRequest, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(db.EmailChangeRequest), args.Error(1)
}

func (m *MockEmailChangeRepository) GetPendingByUserID(ctx context.Context, userID uuid.UUID) (db.EmailChangeRequest, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.EmailChangeRequest), args.Error(1)
}

func (m *MockEmailChangeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEmailChangeRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestRequestEmailChange_Success(t *testing.T) {
	userID := uuid.New()
	currentEmail := "old@example.com"
	newEmail := "new@example.com"
	password := "Password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), service.BcryptCost)

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{
			ID:           userID,
			Email:        currentEmail,
			DisplayName:  "Test User",
			PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
		}, nil)
	userRepo.On("GetUserByEmail", mock.Anything, newEmail).
		Return(db.User{}, sql.ErrNoRows)
	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{}, repository.ErrTokenNotFound)
	changeRepo.On("Create", mock.Anything, userID, newEmail, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(db.EmailChangeRequest{}, nil)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.RequestEmailChange(context.Background(), userID, newEmail, password)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	changeRepo.AssertExpectations(t)
}

func TestRequestEmailChange_UserNotFound(t *testing.T) {
	userID := uuid.New()
	newEmail := "new@example.com"
	password := "Password123"

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{}, sql.ErrNoRows)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.RequestEmailChange(context.Background(), userID, newEmail, password)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrUserNotFound))
	userRepo.AssertExpectations(t)
}

func TestRequestEmailChange_EmailExists(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	currentEmail := "old@example.com"
	newEmail := "new@example.com"
	password := "Password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), service.BcryptCost)

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{
			ID:           userID,
			Email:        currentEmail,
			PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
		}, nil)
	userRepo.On("GetUserByEmail", mock.Anything, newEmail).
		Return(db.User{
			ID:    otherUserID,
			Email: newEmail,
		}, nil)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.RequestEmailChange(context.Background(), userID, newEmail, password)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrEmailExists))
	userRepo.AssertExpectations(t)
}

func TestRequestEmailChange_NoPasswordSet(t *testing.T) {
	userID := uuid.New()
	currentEmail := "old@example.com"
	newEmail := "new@example.com"
	password := "Password123"

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{
			ID:           userID,
			Email:        currentEmail,
			PasswordHash: sql.NullString{Valid: false},
		}, nil)
	userRepo.On("GetUserByEmail", mock.Anything, newEmail).
		Return(db.User{}, sql.ErrNoRows)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.RequestEmailChange(context.Background(), userID, newEmail, password)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrNoPasswordSet))
	userRepo.AssertExpectations(t)
}

func TestRequestEmailChange_WrongPassword(t *testing.T) {
	userID := uuid.New()
	currentEmail := "old@example.com"
	newEmail := "new@example.com"
	password := "Password123"
	wrongPassword := "WrongPassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), service.BcryptCost)

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{
			ID:           userID,
			Email:        currentEmail,
			PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
		}, nil)
	userRepo.On("GetUserByEmail", mock.Anything, newEmail).
		Return(db.User{}, sql.ErrNoRows)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.RequestEmailChange(context.Background(), userID, newEmail, wrongPassword)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrWrongPassword))
	userRepo.AssertExpectations(t)
}

func TestRequestEmailChange_SameEmail(t *testing.T) {
	userID := uuid.New()
	currentEmail := "same@example.com"
	password := "Password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), service.BcryptCost)

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{
			ID:           userID,
			Email:        currentEmail,
			PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
		}, nil)
	userRepo.On("GetUserByEmail", mock.Anything, currentEmail).
		Return(db.User{
			ID:    userID,
			Email: currentEmail,
		}, nil)
	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{}, repository.ErrTokenNotFound)
	changeRepo.On("Create", mock.Anything, userID, currentEmail, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(db.EmailChangeRequest{}, nil)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.RequestEmailChange(context.Background(), userID, currentEmail, password)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	changeRepo.AssertExpectations(t)
}

func TestRequestEmailChange_ReplacePendingRequest(t *testing.T) {
	userID := uuid.New()
	currentEmail := "old@example.com"
	newEmail := "new@example.com"
	password := "Password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), service.BcryptCost)
	pendingID := uuid.New()

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{
			ID:           userID,
			Email:        currentEmail,
			PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
		}, nil)
	userRepo.On("GetUserByEmail", mock.Anything, newEmail).
		Return(db.User{}, sql.ErrNoRows)
	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{
			ID:        pendingID,
			UserID:    userID,
			NewEmail:  "another@example.com",
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)
	changeRepo.On("Delete", mock.Anything, pendingID).Return(nil)
	changeRepo.On("Create", mock.Anything, userID, newEmail, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(db.EmailChangeRequest{}, nil)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.RequestEmailChange(context.Background(), userID, newEmail, password)

	assert.NoError(t, err)
	changeRepo.AssertExpectations(t)
}

func TestVerifyEmailChange_Success(t *testing.T) {
	userID := uuid.New()
	newEmail := "new@example.com"
	token := "valid-token"
	accessToken := "new-access-token"

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("FindValidByTokenHash", mock.Anything, mock.AnythingOfType("string")).
		Return(db.EmailChangeRequest{
			ID:        uuid.New(),
			UserID:    userID,
			NewEmail:  newEmail,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{
			ID:          userID,
			Email:       "old@example.com",
			DisplayName: "Test User",
		}, nil)
	userRepo.On("UpdateUserEmail", mock.Anything, userID, newEmail).Return(nil)
	changeRepo.On("DeleteAllForUser", mock.Anything, userID).Return(nil)
	tokenSvc.On("GenerateAccessToken", userID, newEmail).Return(accessToken, nil)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	result, err := svc.VerifyEmailChange(context.Background(), token)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, accessToken, result.AccessToken)
	assert.Equal(t, newEmail, result.User.Email)
	assert.Equal(t, userID, result.User.ID)
	assert.True(t, result.User.EmailVerified)
	userRepo.AssertExpectations(t)
	changeRepo.AssertExpectations(t)
	tokenSvc.AssertExpectations(t)
}

func TestVerifyEmailChange_InvalidToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("FindValidByTokenHash", mock.Anything, mock.AnythingOfType("string")).
		Return(db.EmailChangeRequest{}, repository.ErrTokenNotFound)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	result, err := svc.VerifyEmailChange(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, repository.ErrTokenNotFound))
	changeRepo.AssertExpectations(t)
}

func TestVerifyEmailChange_ExpiredToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("FindValidByTokenHash", mock.Anything, mock.AnythingOfType("string")).
		Return(db.EmailChangeRequest{}, repository.ErrTokenExpired)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	result, err := svc.VerifyEmailChange(context.Background(), "expired-token")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, repository.ErrTokenExpired))
	changeRepo.AssertExpectations(t)
}

func TestVerifyEmailChange_UserNotFound(t *testing.T) {
	userID := uuid.New()
	newEmail := "new@example.com"

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("FindValidByTokenHash", mock.Anything, mock.AnythingOfType("string")).
		Return(db.EmailChangeRequest{
			ID:        uuid.New(),
			UserID:    userID,
			NewEmail:  newEmail,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{}, sql.ErrNoRows)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	result, err := svc.VerifyEmailChange(context.Background(), "valid-token")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "getting user")
	userRepo.AssertExpectations(t)
	changeRepo.AssertExpectations(t)
}

func TestVerifyEmailChange_UpdateEmailFails(t *testing.T) {
	userID := uuid.New()
	newEmail := "new@example.com"

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("FindValidByTokenHash", mock.Anything, mock.AnythingOfType("string")).
		Return(db.EmailChangeRequest{
			ID:        uuid.New(),
			UserID:    userID,
			NewEmail:  newEmail,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(db.User{
			ID:          userID,
			Email:       "old@example.com",
			DisplayName: "Test User",
		}, nil)
	userRepo.On("UpdateUserEmail", mock.Anything, userID, newEmail).Return(errors.New("update failed"))

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	result, err := svc.VerifyEmailChange(context.Background(), "valid-token")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "updating user email")
	userRepo.AssertExpectations(t)
	changeRepo.AssertExpectations(t)
}

func TestCancelEmailChange_Success(t *testing.T) {
	userID := uuid.New()
	pendingID := uuid.New()

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{
			ID:        pendingID,
			UserID:    userID,
			NewEmail:  "new@example.com",
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)
	changeRepo.On("Delete", mock.Anything, pendingID).Return(nil)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.CancelEmailChange(context.Background(), userID)

	assert.NoError(t, err)
	changeRepo.AssertExpectations(t)
}

func TestCancelEmailChange_NoPendingRequest(t *testing.T) {
	userID := uuid.New()

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{}, repository.ErrTokenNotFound)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.CancelEmailChange(context.Background(), userID)

	assert.NoError(t, err)
	changeRepo.AssertExpectations(t)
}

func TestCancelEmailChange_ExpiredRequest(t *testing.T) {
	userID := uuid.New()

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{}, repository.ErrTokenExpired)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	err := svc.CancelEmailChange(context.Background(), userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "finding pending request")
	changeRepo.AssertExpectations(t)
}

func TestGetPendingEmailChange_Success(t *testing.T) {
	userID := uuid.New()
	newEmail := "new@example.com"
	expiresAt := time.Now().Add(time.Hour)

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{
			ID:        uuid.New(),
			UserID:    userID,
			NewEmail:  newEmail,
			ExpiresAt: expiresAt,
		}, nil)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	email, gotExpiresAt, err := svc.GetPendingEmailChange(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, newEmail, email)
	assert.Equal(t, expiresAt, gotExpiresAt)
	changeRepo.AssertExpectations(t)
}

func TestGetPendingEmailChange_NotFound(t *testing.T) {
	userID := uuid.New()

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{}, repository.ErrTokenNotFound)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	email, _, err := svc.GetPendingEmailChange(context.Background(), userID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrTokenNotFound))
	assert.Empty(t, email)
	changeRepo.AssertExpectations(t)
}

func TestGetPendingEmailChange_Expired(t *testing.T) {
	userID := uuid.New()

	userRepo := new(MockUserRepository)
	changeRepo := new(MockEmailChangeRepository)
	tokenSvc := new(MockTokenService)

	changeRepo.On("GetPendingByUserID", mock.Anything, userID).
		Return(db.EmailChangeRequest{}, repository.ErrTokenExpired)

	svc := service.NewEmailChangeService(userRepo, changeRepo, tokenSvc, "https://example.com")

	email, _, err := svc.GetPendingEmailChange(context.Background(), userID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrTokenExpired))
	assert.Empty(t, email)
	changeRepo.AssertExpectations(t)
}
