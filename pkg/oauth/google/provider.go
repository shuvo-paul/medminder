// Package google provides OAuth 2.0 authentication with Google.
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/shuvo-paul/medminder/pkg/oauth"
)

// Google OAuth 2.0 endpoints.
var (
	AuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	TokenURL    = "https://oauth2.googleapis.com/token"
	UserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// GoogleProvider implements the oauth.Provider interface for Google.
type GoogleProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// New creates a new Google OAuth provider.
func New() (*GoogleProvider, error) {
	return &GoogleProvider{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("GOOGLE_REDIRECT_URI"),
	}, nil
}

// Name returns the provider identifier.
func (p *GoogleProvider) Name() string {
	return "google"
}

// RequiredEnvVars returns the required environment variables.
func (p *GoogleProvider) RequiredEnvVars() []string {
	return []string{"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "GOOGLE_REDIRECT_URI"}
}

// AuthURL returns the Google authorization URL.
func (p *GoogleProvider) AuthURL(state string) string {
	params := url.Values{
		"client_id":     {p.ClientID},
		"redirect_uri":  {p.RedirectURI},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
	}
	return AuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens.
func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*oauth.TokenResponse, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"redirect_uri":  {p.RedirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("google: creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google: exchanging code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("google: token exchange failed: %s - %s", errResp.Error, errResp.ErrorDescription)
	}

	var tokenResp oauth.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("google: decoding token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo fetches user information from Google.
func (p *GoogleProvider) GetUserInfo(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("google: creating userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google: getting user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google: userinfo request failed with status %d", resp.StatusCode)
	}

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"verified_email"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("google: decoding user info: %w", err)
	}

	return &oauth.UserInfo{
		ProviderUserID: userInfo.ID,
		Email:          userInfo.Email,
		EmailVerified:  userInfo.EmailVerified,
		Name:           userInfo.Name,
	}, nil
}

func init() {
	// Auto-register the Google provider if environment variables are set
	provider, err := New()
	if err != nil {
		return // Should not happen since New() doesn't return error
	}

	// Check if required env vars are present
	missing := false
	for _, envVar := range provider.RequiredEnvVars() {
		if os.Getenv(envVar) == "" {
			missing = true
			break
		}
	}

	if !missing {
		if err := oauth.Register(provider); err != nil {
			fmt.Printf("oauth/google: failed to register provider: %v\n", err)
		}
	}
}

// TokenResponseWithExpiry includes expiry time for caching.
type TokenResponseWithExpiry struct {
	oauth.TokenResponse
	Expiry time.Time `json:"expiry"`
}

// IsExpired checks if the token has expired.
func (t *TokenResponseWithExpiry) IsExpired() bool {
	return time.Now().After(t.Expiry)
}

// NewTokenResponseWithExpiry creates a TokenResponseWithExpiry from a TokenResponse.
func NewTokenResponseWithExpiry(resp *oauth.TokenResponse) *TokenResponseWithExpiry {
	return &TokenResponseWithExpiry{
		TokenResponse: *resp,
		Expiry:        time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}
}
