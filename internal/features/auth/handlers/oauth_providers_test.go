package handlers_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	db "github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/pkg/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListOAuthProvidersHandler(t *testing.T) {
	// Set required environment variables for the mock provider
	os.Setenv("TEST_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_CLIENT_SECRET", "test-client-secret")
	os.Setenv("TEST_REDIRECT_URI", "http://localhost:8080/callback")
	defer os.Unsetenv("TEST_CLIENT_ID")
	defer os.Unsetenv("TEST_CLIENT_SECRET")
	defer os.Unsetenv("TEST_REDIRECT_URI")

	// Register a mock provider for testing
	mockProvider := &oauth.MockProvider{
		NameVal:            "test-provider",
		RequiredEnvVarsVal: []string{"TEST_CLIENT_ID", "TEST_CLIENT_SECRET", "TEST_REDIRECT_URI"},
		AuthURLVal:         "https://test-provider.example.com/auth",
	}
	err := oauth.Register(mockProvider)
	require.NoError(t, err)

	handler := handlers.ListOAuthProvidersHandler()

	// Call handler directly
	resp, err := handler(context.Background(), &struct{}{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Providers)

	// Find our test provider
	var found bool
	for _, p := range resp.Providers {
		if p.ID == "test-provider" {
			found = true
			assert.Equal(t, "Test-Provider", p.Name) // strings.Title capitalizes each word
			assert.Equal(t, "/assets/icons/test-provider.svg", p.IconURL)
			assert.Equal(t, "/api/auth/oauth/test-provider/callback", p.CallbackPath)
			break
		}
	}
	assert.True(t, found, "test-provider should be in the list")
}

// encodeOAuthState creates a properly encoded OAuthState string for testing
func encodeOAuthState(nonce, redirect, purpose string) string {
	state := dto.OAuthState{
		Nonce:    nonce,
		Redirect: redirect,
		Purpose:  purpose,
	}
	data, _ := json.Marshal(state)
	return base64.URLEncoding.EncodeToString(data)
}

// newMockAuditRepo creates a MockAuditRepository pre-configured to accept any LogEvent call.
func newMockAuditRepo() *MockAuditRepository {
	m := new(MockAuditRepository)
	m.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return m
}

// newCallbackDeps creates dependencies for testing the callback handler.
func newCallbackDeps() *handlers.OAuthHandlerDeps {
	return &handlers.OAuthHandlerDeps{
		AuthCodeRepo: &MockOAuthAuthorizationCodeRepository{},
		OAuthSvc:     &MockOAuthService{},
		TokenSvc:     &MockTokenService{},
		TokenRepo:    &MockRefreshTokenRepository{},
		AuditRepo:    newMockAuditRepo(),
		FrontendURL:  "http://localhost:5173",
	}
}

func TestInitiateOAuthHandler_Success(t *testing.T) {
	// Set required environment variables for the mock provider
	os.Setenv("TEST_CLIENT_ID", "test-client-id")
	os.Setenv("TEST_CLIENT_SECRET", "test-client-secret")
	os.Setenv("TEST_REDIRECT_URI", "http://localhost:8080/callback")
	defer os.Unsetenv("TEST_CLIENT_ID")
	defer os.Unsetenv("TEST_CLIENT_SECRET")
	defer os.Unsetenv("TEST_REDIRECT_URI")

	// Register a mock provider for testing
	mockProvider := &oauth.MockProvider{
		NameVal:            "test-provider",
		RequiredEnvVarsVal: []string{"TEST_CLIENT_ID", "TEST_CLIENT_SECRET", "TEST_REDIRECT_URI"},
		AuthURLVal:         "https://test-provider.example.com/auth",
	}
	_ = oauth.Register(mockProvider)

	handler := handlers.InitiateOAuthHandler()

	// Create input with provider and properly encoded state
	state := encodeOAuthState("test-nonce-123", "/dashboard", "login")
	input := &dto.InitiateOAuthInput{
		Provider: "test-provider",
		State:    state,
	}

	// Call handler
	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://test-provider.example.com/auth", resp.Location)
}

func TestInitiateOAuthHandler_MissingState(t *testing.T) {
	handler := handlers.InitiateOAuthHandler()

	input := &dto.InitiateOAuthInput{
		Provider: "google",
		State:    "",
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "state parameter is required")
}

func TestInitiateOAuthHandler_InvalidState(t *testing.T) {
	handler := handlers.InitiateOAuthHandler()

	input := &dto.InitiateOAuthInput{
		Provider: "google",
		State:    "not-valid-base64!!!",
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid state parameter")
}

func TestInitiateOAuthHandler_OpenRedirect(t *testing.T) {
	handler := handlers.InitiateOAuthHandler()

	// Absolute URL with scheme — must be rejected as open redirect.
	state := encodeOAuthState("test-nonce", "https://evil.example.com/steal", "login")
	input := &dto.InitiateOAuthInput{
		Provider: "google",
		State:    state,
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid redirect URL")
}

func TestInitiateOAuthHandler_ProviderNotFound(t *testing.T) {
	handler := handlers.InitiateOAuthHandler()

	state := encodeOAuthState("test-nonce", "/dashboard", "login")
	input := &dto.InitiateOAuthInput{
		Provider: "unknown-provider",
		State:    state,
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestOAuthCallbackHandler_ProviderError_WithState(t *testing.T) {
	deps := newCallbackDeps()
	handler := handlers.OAuthCallbackHandler(deps)

	state := encodeOAuthState("test-nonce", "/dashboard", "login")
	input := &dto.OAuthCallbackInput{
		Provider: "google",
		Error:    "access_denied",
		State:    state,
	}

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "http://localhost:5173/dashboard?oauth_error=cancelled&provider=google", resp.Redirect)
}

func TestOAuthCallbackHandler_ProviderError_EmptyState(t *testing.T) {
	deps := newCallbackDeps()
	handler := handlers.OAuthCallbackHandler(deps)

	input := &dto.OAuthCallbackInput{
		Provider: "google",
		Error:    "access_denied",
		State:    "",
	}

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "http://localhost:5173/login?oauth_error=cancelled", resp.Redirect)
}

func TestOAuthCallbackHandler_ProviderError_MalformedState(t *testing.T) {
	deps := newCallbackDeps()
	handler := handlers.OAuthCallbackHandler(deps)

	input := &dto.OAuthCallbackInput{
		Provider: "google",
		Error:    "access_denied",
		State:    "not-valid-base64!!!",
	}

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "http://localhost:5173/login?oauth_error=cancelled", resp.Redirect)
}

func TestOAuthCallbackHandler_MissingCode(t *testing.T) {
	deps := newCallbackDeps()
	handler := handlers.OAuthCallbackHandler(deps)

	state := encodeOAuthState("test-nonce", "/dashboard", "login")
	input := &dto.OAuthCallbackInput{
		Provider: "google",
		State:    state,
		// No Code, No Error
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	require.Contains(t, err.Error(), "code")
}

func TestOAuthCallbackHandler_Success(t *testing.T) {
	// Set up environment and register a mock provider
	os.Setenv("TEST_CALLBACK_ID", "test-id")
	os.Setenv("TEST_CALLBACK_SECRET", "test-secret")
	os.Setenv("TEST_CALLBACK_REDIRECT", "http://localhost:8080/callback")
	defer func() {
		os.Unsetenv("TEST_CALLBACK_ID")
		os.Unsetenv("TEST_CALLBACK_SECRET")
		os.Unsetenv("TEST_CALLBACK_REDIRECT")
	}()

	mockProvider := &oauth.MockProvider{
		NameVal:            "google",
		RequiredEnvVarsVal: []string{"TEST_CALLBACK_ID", "TEST_CALLBACK_SECRET", "TEST_CALLBACK_REDIRECT"},
		ExchangeCodeResp: &oauth.TokenResponse{
			AccessToken: "google-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		},
		GetUserInfoResp: &oauth.UserInfo{
			ProviderUserID: "google-user-123",
			Email:          "test@example.com",
			EmailVerified:  true,
			Name:           "Test User",
		},
	}
	_ = oauth.Register(mockProvider)

	// Set up mock authorization code repository
	mockCodeRepo := new(MockOAuthAuthorizationCodeRepository)
	mockCodeRepo.On("CreateAuthorizationCodeWithUserInfo",
		mock.Anything,
		mock.Anything, // id
		mock.Anything, // codeHash
		"test-nonce",  // nonce
		"login",       // purpose
		mock.Anything, // expiresAt
		"google",      // provider
		"google-user-123",
		"test@example.com",
		"Test User",
		true,
	).Return(db.OauthAuthorizationCode{}, nil)

	deps := &handlers.OAuthHandlerDeps{
		AuthCodeRepo: mockCodeRepo,
		OAuthSvc:     &MockOAuthService{},
		TokenSvc:     &MockTokenService{},
		TokenRepo:    &MockRefreshTokenRepository{},
		AuditRepo:    newMockAuditRepo(),
		FrontendURL:  "http://localhost:5173",
	}
	handler := handlers.OAuthCallbackHandler(deps)

	state := encodeOAuthState("test-nonce", "/dashboard", "login")
	input := &dto.OAuthCallbackInput{
		Provider: "google",
		Code:     "google-auth-code",
		State:    state,
	}

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Redirect, "http://localhost:5173/auth/callback?code=")
	assert.Contains(t, resp.Redirect, "&state=")
	// No JWT tokens in the redirect URL
	assert.NotContains(t, resp.Redirect, "access_token")
	assert.NotContains(t, resp.Redirect, "refresh_token")
}

// -- Token Exchange Handler Tests --

func TestTokenExchangeHandler_Success(t *testing.T) {
	mockCodeRepo := new(MockOAuthAuthorizationCodeRepository)
	mockOAuthSvc := new(MockOAuthService)
	mockTokenSvc := new(MockTokenService)
	mockTokenRepo := new(MockRefreshTokenRepository)

	userID := uuid.New()

	// Mock authorization code lookup
	authCodeInfo := &repository.AuthorizationCodeInfo{
		OauthAuthorizationCode: db.OauthAuthorizationCode{
			CodeHash: "test-code-hash",
			Nonce:    "test-nonce-123",
			Purpose:  "login",
		},
		Provider:              "google",
		ProviderUserID:        "google-user-123",
		ProviderEmail:         "test@example.com",
		ProviderName:          "Test User",
		ProviderEmailVerified: true,
	}
	mockCodeRepo.On("GetAndLockAuthorizationCode", mock.Anything, mock.Anything).Return(authCodeInfo, nil)
	mockCodeRepo.On("MarkAuthorizationCodeAsUsed", mock.Anything, mock.Anything).Return(db.OauthAuthorizationCode{}, nil)

	// Mock OAuth service
	mockOAuthSvc.On("GetOrCreateUserByOAuth", mock.Anything, "google", mock.Anything).Return(&service.OAuthUser{
		ID:            userID,
		Email:         "test@example.com",
		DisplayName:   "Test User",
		EmailVerified: true,
	}, nil)

	// Mock token generation
	mockTokenSvc.On("GenerateAccessToken", userID, "test@example.com").Return("access-token-xyz", nil)
	mockTokenSvc.On("GenerateRefreshToken").Return("refresh-token-xyz", nil)
	mockTokenSvc.On("HashRefreshToken", "refresh-token-xyz").Return("hashed-refresh-token")
	mockTokenRepo.On("CreateRefreshToken", mock.Anything, userID, "hashed-refresh-token", mock.Anything).Return(db.CreateRefreshTokenRow{}, nil)

	deps := &handlers.OAuthHandlerDeps{
		AuthCodeRepo: mockCodeRepo,
		OAuthSvc:     mockOAuthSvc,
		TokenSvc:     mockTokenSvc,
		TokenRepo:    mockTokenRepo,
	}
	handler := handlers.TokenExchangeHandler(deps)

	// Generate a proper code hash for the state nonce to match
	rawCode := "raw-internal-code"

	state := encodeOAuthState("test-nonce-123", "/dashboard", "login")
	input := &dto.OAuthTokenRequest{
		Body: struct {
			Code  string `json:"code" maxLength:"256"`
			State string `json:"state" maxLength:"1024"`
		}{
			Code:  rawCode,
			State: state,
		},
	}

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "access-token-xyz", resp.Body.AccessToken)
	assert.Equal(t, "refresh-token-xyz", resp.Body.RefreshToken)
	assert.Equal(t, "Bearer", resp.Body.TokenType)
	assert.Equal(t, int(service.AccessTokenExpiry.Seconds()), resp.Body.ExpiresIn)
	assert.Equal(t, userID.String(), resp.Body.User.ID)
	assert.Equal(t, "test@example.com", resp.Body.User.Email)
}

func TestTokenExchangeHandler_InvalidCode(t *testing.T) {
	mockCodeRepo := new(MockOAuthAuthorizationCodeRepository)
	mockCodeRepo.On("GetAndLockAuthorizationCode", mock.Anything, mock.Anything).Return(nil, repository.ErrOAuthCodeNotFound)

	deps := &handlers.OAuthHandlerDeps{
		AuthCodeRepo: mockCodeRepo,
		OAuthSvc:     &MockOAuthService{},
		TokenSvc:     &MockTokenService{},
		TokenRepo:    &MockRefreshTokenRepository{},
		AuditRepo:    newMockAuditRepo(),
	}
	handler := handlers.TokenExchangeHandler(deps)

	state := encodeOAuthState("test-nonce", "/dashboard", "login")
	input := &dto.OAuthTokenRequest{
		Body: struct {
			Code  string `json:"code" maxLength:"256"`
			State string `json:"state" maxLength:"1024"`
		}{
			Code:  "invalid-code",
			State: state,
		},
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), string(dto.OAuthErrorInvalidCode))
	auditRepo := deps.AuditRepo.(*MockAuditRepository)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_code_rejected", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTokenExchangeHandler_NonceMismatch(t *testing.T) {
	mockCodeRepo := new(MockOAuthAuthorizationCodeRepository)

	authCodeInfo := &repository.AuthorizationCodeInfo{
		OauthAuthorizationCode: db.OauthAuthorizationCode{
			CodeHash: "test-code-hash",
			Nonce:    "different-nonce", // Different from state
			Purpose:  "login",
		},
		Provider:              "google",
		ProviderUserID:        "google-user-123",
		ProviderEmail:         "test@example.com",
		ProviderName:          "Test User",
		ProviderEmailVerified: true,
	}
	mockCodeRepo.On("GetAndLockAuthorizationCode", mock.Anything, mock.Anything).Return(authCodeInfo, nil)
	// Nonce mismatch marks code as used
	mockCodeRepo.On("MarkAuthorizationCodeAsUsed", mock.Anything, mock.Anything).Return(db.OauthAuthorizationCode{}, nil)

	deps := &handlers.OAuthHandlerDeps{
		AuthCodeRepo: mockCodeRepo,
		OAuthSvc:     &MockOAuthService{},
		TokenSvc:     &MockTokenService{},
		TokenRepo:    &MockRefreshTokenRepository{},
		AuditRepo:    newMockAuditRepo(),
	}
	handler := handlers.TokenExchangeHandler(deps)

	state := encodeOAuthState("correct-nonce", "/dashboard", "login")
	input := &dto.OAuthTokenRequest{
		Body: struct {
			Code  string `json:"code" maxLength:"256"`
			State string `json:"state" maxLength:"1024"`
		}{
			Code:  "some-code",
			State: state,
		},
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), string(dto.OAuthErrorInvalidCode))
	mockCodeRepo.AssertCalled(t, "MarkAuthorizationCodeAsUsed", mock.Anything, mock.Anything)
	auditRepo := deps.AuditRepo.(*MockAuditRepository)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_code_rejected", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTokenExchangeHandler_EmailExists(t *testing.T) {
	mockCodeRepo := new(MockOAuthAuthorizationCodeRepository)
	mockOAuthSvc := new(MockOAuthService)

	authCodeInfo := &repository.AuthorizationCodeInfo{
		OauthAuthorizationCode: db.OauthAuthorizationCode{
			CodeHash: "test-code-hash",
			Nonce:    "test-nonce",
			Purpose:  "login",
		},
		Provider:              "google",
		ProviderUserID:        "google-user-123",
		ProviderEmail:         "exists@example.com",
		ProviderName:          "Existing User",
		ProviderEmailVerified: true,
	}
	mockCodeRepo.On("GetAndLockAuthorizationCode", mock.Anything, mock.Anything).Return(authCodeInfo, nil)
	mockCodeRepo.On("MarkAuthorizationCodeAsUsed", mock.Anything, mock.Anything).Return(db.OauthAuthorizationCode{}, nil)

	mockOAuthSvc.On("GetOrCreateUserByOAuth", mock.Anything, "google", mock.Anything).Return(nil, service.ErrEmailExists)

	deps := &handlers.OAuthHandlerDeps{
		AuthCodeRepo: mockCodeRepo,
		OAuthSvc:     mockOAuthSvc,
		TokenSvc:     &MockTokenService{},
		TokenRepo:    &MockRefreshTokenRepository{},
		AuditRepo:    newMockAuditRepo(),
	}
	handler := handlers.TokenExchangeHandler(deps)

	state := encodeOAuthState("test-nonce", "/dashboard", "login")
	input := &dto.OAuthTokenRequest{
		Body: struct {
			Code  string `json:"code" maxLength:"256"`
			State string `json:"state" maxLength:"1024"`
		}{
			Code:  "some-code",
			State: state,
		},
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), string(dto.OAuthErrorEmailExists))
	mockCodeRepo.AssertCalled(t, "MarkAuthorizationCodeAsUsed", mock.Anything, mock.Anything)
	auditRepo := deps.AuditRepo.(*MockAuditRepository)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_login_failed", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTokenExchangeHandler_MalformedState(t *testing.T) {
	deps := &handlers.OAuthHandlerDeps{
		AuthCodeRepo: &MockOAuthAuthorizationCodeRepository{},
		OAuthSvc:     &MockOAuthService{},
		TokenSvc:     &MockTokenService{},
		TokenRepo:    &MockRefreshTokenRepository{},
		AuditRepo:    newMockAuditRepo(),
	}
	handler := handlers.TokenExchangeHandler(deps)

	input := &dto.OAuthTokenRequest{
		Body: struct {
			Code  string `json:"code" maxLength:"256"`
			State string `json:"state" maxLength:"1024"`
		}{
			Code:  "some-code",
			State: "not-valid-base64!!!",
		},
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestTokenExchangeHandler_EmptyState(t *testing.T) {
	deps := &handlers.OAuthHandlerDeps{
		AuthCodeRepo: &MockOAuthAuthorizationCodeRepository{},
		OAuthSvc:     &MockOAuthService{},
		TokenSvc:     &MockTokenService{},
		TokenRepo:    &MockRefreshTokenRepository{},
		AuditRepo:    newMockAuditRepo(),
	}
	handler := handlers.TokenExchangeHandler(deps)

	input := &dto.OAuthTokenRequest{
		Body: struct {
			Code  string `json:"code" maxLength:"256"`
			State string `json:"state" maxLength:"1024"`
		}{
			Code:  "some-code",
			State: "",
		},
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// -- OAuth Link Init Handler Tests --

func newLinkInitDeps() *handlers.OAuthHandlerDeps {
	return &handlers.OAuthHandlerDeps{
		AuthCodeRepo: &MockOAuthAuthorizationCodeRepository{},
		OAuthSvc:     &MockOAuthService{},
		TokenSvc:     &MockTokenService{},
		TokenRepo:    &MockRefreshTokenRepository{},
		AuditRepo:    newMockAuditRepo(),
	}
}

func TestOAuthLinkInitHandler_Success(t *testing.T) {
	// Register a mock provider for link init to verify against
	os.Setenv("LINK_TEST_ID", "test-id")
	os.Setenv("LINK_TEST_SECRET", "test-secret")
	os.Setenv("LINK_TEST_REDIRECT", "http://localhost:8080/callback")
	defer func() {
		os.Unsetenv("LINK_TEST_ID")
		os.Unsetenv("LINK_TEST_SECRET")
		os.Unsetenv("LINK_TEST_REDIRECT")
	}()

	mockProvider := &oauth.MockProvider{
		NameVal:            "http://link.test",
		RequiredEnvVarsVal: []string{"LINK_TEST_ID", "LINK_TEST_SECRET", "LINK_TEST_REDIRECT"},
	}
	err := oauth.Register(mockProvider)
	require.NoError(t, err, "failed to register mock provider")

	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	deps := newLinkInitDeps()
	mockTokenSvc := deps.TokenSvc.(*MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockCodeRepo := deps.AuthCodeRepo.(*MockOAuthAuthorizationCodeRepository)
	mockCodeRepo.On("CreateAuthorizationCode",
		mock.Anything,
		mock.AnythingOfType("uuid.UUID"),     // id
		mock.AnythingOfType("string"),        // codeHash
		mock.AnythingOfType("uuid.NullUUID"), // userID
		mock.AnythingOfType("string"),        // nonce
		"link",                               // purpose
		mock.AnythingOfType("time.Time"),     // expiresAt
	).Return(db.OauthAuthorizationCode{}, nil)

	handler := handlers.OAuthLinkInitHandler(deps)

	input := &dto.OAuthLinkInitInput{
		Authorization: "Bearer " + token,
		Provider:      "http://link.test",
	}

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify state contains purpose=link
	state, err := dto.ParseOAuthState(resp.Body.State)
	require.NoError(t, err)
	assert.Equal(t, "link", state.Purpose)
	assert.NotEmpty(t, state.Nonce)
}

func TestOAuthLinkInitHandler_MissingAuth(t *testing.T) {
	deps := newLinkInitDeps()
	handler := handlers.OAuthLinkInitHandler(deps)

	input := &dto.OAuthLinkInitInput{
		Authorization: "",
		Provider:      "google",
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired access token")
}

func TestOAuthLinkInitHandler_InvalidToken(t *testing.T) {
	deps := newLinkInitDeps()
	mockTokenSvc := deps.TokenSvc.(*MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", "bad-token").Return(nil, service.ErrInvalidToken)

	handler := handlers.OAuthLinkInitHandler(deps)

	input := &dto.OAuthLinkInitInput{
		Authorization: "Bearer bad-token",
		Provider:      "google",
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired")
}

func TestOAuthLinkInitHandler_ProviderNotFound(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	deps := newLinkInitDeps()
	mockTokenSvc := deps.TokenSvc.(*MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	handler := handlers.OAuthLinkInitHandler(deps)

	input := &dto.OAuthLinkInitInput{
		Authorization: "Bearer " + token,
		Provider:      "unknown-provider",
	}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "provider not found")
}
