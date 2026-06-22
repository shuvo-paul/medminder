package middleware_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/middleware"
	"github.com/stretchr/testify/assert"
)

type mockHumaContext struct {
	ctx          context.Context
	params       map[string]string
	headers      map[string]string
	statusCode   int
	responseBody bytes.Buffer
	responseType string
	nextCalled   bool
}

func (m *mockHumaContext) Operation() *huma.Operation                 { return nil }
func (m *mockHumaContext) Version() huma.ProtoVersion                 { return huma.ProtoVersion{} }
func (m *mockHumaContext) Context() context.Context                   { return m.ctx }
func (m *mockHumaContext) TLS() *tls.ConnectionState                  { return nil }
func (m *mockHumaContext) Method() string                             { return "GET" }
func (m *mockHumaContext) Host() string                               { return "localhost" }
func (m *mockHumaContext) RemoteAddr() string                         { return "127.0.0.1" }
func (m *mockHumaContext) URL() url.URL                               { return url.URL{Path: "/api/profiles/123"} }
func (m *mockHumaContext) Param(name string) string                   { return m.params[name] }
func (m *mockHumaContext) Query(name string) string                   { return "" }
func (m *mockHumaContext) Header(name string) string                  { return m.headers[name] }
func (m *mockHumaContext) EachHeader(fn func(name, value string))     {}
func (m *mockHumaContext) BodyReader() io.Reader                      { return ioutil.NopCloser(bytes.NewReader(nil)) }
func (m *mockHumaContext) GetMultipartForm() (*multipart.Form, error) { return nil, nil }
func (m *mockHumaContext) SetReadDeadline(t time.Time) error          { return nil }
func (m *mockHumaContext) SetStatus(code int)                         { m.statusCode = code }
func (m *mockHumaContext) Status() int                                { return m.statusCode }
func (m *mockHumaContext) SetHeader(name, value string)               { m.responseType = value }
func (m *mockHumaContext) AppendHeader(name, value string)            {}
func (m *mockHumaContext) BodyWriter() io.Writer                      { return &m.responseBody }

type mockTokenValidator struct {
	claims jwt.MapClaims
	err    error
}

func (m *mockTokenValidator) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	return m.claims, m.err
}

type mockPermissionChecker struct {
	allowed bool
	err     error
}

func (m *mockPermissionChecker) HasPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permission string) (bool, error) {
	return m.allowed, m.err
}

func (m *mockPermissionChecker) HasAnyPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permissions []string) (bool, error) {
	return m.allowed, m.err
}

func TestHumaRequireProfilePermission_Allowed(t *testing.T) {
	userID := uuid.New()
	mockCtx := &mockHumaContext{
		ctx:     context.Background(),
		params:  map[string]string{"id": uuid.New().String()},
		headers: map[string]string{"Authorization": "Bearer valid-token"},
	}
	nextCalled := false

	middleware.HumaRequireProfilePermission(
		&mockPermissionChecker{allowed: true},
		&mockTokenValidator{
			claims: jwt.MapClaims{"sub": userID.String()},
		},
		"profile:read",
	)(mockCtx, func(ctx huma.Context) {
		nextCalled = true
		assert.Equal(t, userID, middleware.UserIDFromContext(ctx.Context()), "userID should be stored in context")
	})

	assert.True(t, nextCalled, "next should be called when permission is granted")
	assert.Equal(t, 0, mockCtx.statusCode, "status should not be set when allowed")
}

func TestHumaRequireProfilePermission_Denied(t *testing.T) {
	userID := uuid.New()
	mockCtx := &mockHumaContext{
		ctx:     context.Background(),
		params:  map[string]string{"id": uuid.New().String()},
		headers: map[string]string{"Authorization": "Bearer valid-token"},
	}
	nextCalled := false

	middleware.HumaRequireProfilePermission(
		&mockPermissionChecker{allowed: false},
		&mockTokenValidator{
			claims: jwt.MapClaims{"sub": userID.String()},
		},
		"profile:read",
	)(mockCtx, func(ctx huma.Context) {
		nextCalled = true
	})

	assert.False(t, nextCalled, "next should not be called when permission is denied")
	assert.Equal(t, 403, mockCtx.statusCode)

	var resp map[string]interface{}
	json.Unmarshal(mockCtx.responseBody.Bytes(), &resp)
	assert.Equal(t, "Forbidden", resp["title"])
}

func TestHumaRequireProfilePermission_NoProfileID(t *testing.T) {
	mockCtx := &mockHumaContext{
		ctx:     context.Background(),
		params:  map[string]string{},
		headers: map[string]string{"Authorization": "Bearer valid-token"},
	}
	nextCalled := false

	middleware.HumaRequireProfilePermission(
		&mockPermissionChecker{allowed: false},
		&mockTokenValidator{},
		"profile:read",
	)(mockCtx, func(ctx huma.Context) {
		nextCalled = true
	})

	assert.True(t, nextCalled, "next should be called when no profile ID (listing endpoint)")
}

func TestHumaRequireProfilePermission_InvalidProfileID(t *testing.T) {
	mockCtx := &mockHumaContext{
		ctx:     context.Background(),
		params:  map[string]string{"id": "not-a-uuid"},
		headers: map[string]string{"Authorization": "Bearer valid-token"},
	}
	nextCalled := false

	middleware.HumaRequireProfilePermission(
		&mockPermissionChecker{},
		&mockTokenValidator{
			claims: jwt.MapClaims{"sub": uuid.New().String()},
		},
		"profile:read",
	)(mockCtx, func(ctx huma.Context) {
		nextCalled = true
	})

	assert.False(t, nextCalled)
	assert.Equal(t, 400, mockCtx.statusCode)
}

func TestHumaRequireProfilePermission_InvalidToken(t *testing.T) {
	mockCtx := &mockHumaContext{
		ctx:     context.Background(),
		params:  map[string]string{"id": uuid.New().String()},
		headers: map[string]string{"Authorization": "Bearer invalid-token"},
	}
	nextCalled := false

	middleware.HumaRequireProfilePermission(
		&mockPermissionChecker{},
		&mockTokenValidator{err: assert.AnError},
		"profile:read",
	)(mockCtx, func(ctx huma.Context) {
		nextCalled = true
	})

	assert.False(t, nextCalled)
	assert.Equal(t, 401, mockCtx.statusCode)
}

func TestHumaRequireProfilePermission_MissingAuthHeader(t *testing.T) {
	mockCtx := &mockHumaContext{
		ctx:     context.Background(),
		params:  map[string]string{"id": uuid.New().String()},
		headers: map[string]string{},
	}
	nextCalled := false

	middleware.HumaRequireProfilePermission(
		&mockPermissionChecker{},
		&mockTokenValidator{},
		"profile:read",
	)(mockCtx, func(ctx huma.Context) {
		nextCalled = true
	})

	assert.False(t, nextCalled)
	assert.Equal(t, 401, mockCtx.statusCode)
}

func TestHumaRequireProfilePermission_AnyOfPermissions(t *testing.T) {
	userID := uuid.New()
	mockCtx := &mockHumaContext{
		ctx:     context.Background(),
		params:  map[string]string{"id": uuid.New().String()},
		headers: map[string]string{"Authorization": "Bearer valid-token"},
	}
	nextCalled := false

	middleware.HumaRequireProfilePermission(
		&mockPermissionChecker{allowed: true},
		&mockTokenValidator{
			claims: jwt.MapClaims{"sub": userID.String()},
		},
		"profile:admin", "profile:owner",
	)(mockCtx, func(ctx huma.Context) {
		nextCalled = true
	})

	assert.True(t, nextCalled, "next should be called when user has any of the required permissions")
}
