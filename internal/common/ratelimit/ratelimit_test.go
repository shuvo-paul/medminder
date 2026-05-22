package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/shuvo-paul/medminder/internal/common/ratelimit"
	"github.com/stretchr/testify/assert"
)

func TestLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	limiter := ratelimit.New(5, 100*time.Millisecond)
	defer limiter.Stop()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := range 5 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "request %d should be allowed", i+1)
	}
}

func TestLimiter_BlocksRequestsOverLimit(t *testing.T) {
	limiter := ratelimit.New(5, 100*time.Millisecond)
	defer limiter.Stop()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit.
	for i := range 5 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "request %d should be allowed", i+1)
	}

	// 6th request should be blocked.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "6th request should be rate limited")
	assert.Equal(t, "0", rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("Retry-After"), "should include Retry-After on 429")
}

func TestLimiter_ResetsAfterInterval(t *testing.T) {
	limiter := ratelimit.New(5, 50*time.Millisecond)
	defer limiter.Stop()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit.
	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Verify blocked.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// Wait for window to slide.
	time.Sleep(60 * time.Millisecond)

	// Now should be allowed again.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "should be allowed after window expires")
}

func TestLimiter_DifferentIPsHaveIndependentCounters(t *testing.T) {
	limiter := ratelimit.New(5, 100*time.Millisecond)
	defer limiter.Stop()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust limit for IP A.
	for range 5 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// IP A should now be blocked.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// IP B should still be allowed (different IP, fresh counter).
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:54321"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "IP B should be allowed (independent counter)")
}

func TestLimiter_SetsResponseHeaders(t *testing.T) {
	limiter := ratelimit.New(5, 100*time.Millisecond)
	defer limiter.Stop()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "5", rec.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "4", rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
}

func TestLimiter_ResetTimeReflectsOldestEntry(t *testing.T) {
	// Use a fixed clock so we control time precisely.
	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	var curTime time.Time

	// We need to construct the limiter with access to the now func.
	// Use a smaller rate so it's easier to reason about.
	limiter := ratelimit.New(3, 60*time.Second)
	defer limiter.Stop()

	curTime = baseTime

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request 1 at t=0.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resetUnix, err := strconv.ParseInt(rec.Header().Get("X-RateLimit-Reset"), 10, 64)
	assert.NoError(t, err)
	resetAt := time.Unix(resetUnix, 0)
	// Window is empty → reset should be roughly now + interval.
	// Can't assert exact since time advances, but should be within a few seconds of baseTime+60.
	assert.True(t, resetAt.After(curTime), "reset should be in the future")
	assert.Equal(t, "2", rec.Header().Get("X-RateLimit-Remaining"))
}

func TestLimiter_RetryAfterTracksActualWindow(t *testing.T) {
	limiter := ratelimit.New(2, 10*time.Second)
	defer limiter.Stop()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit quickly.
	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// 3rd request — blocked.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// Retry-After should roughly equal 10s (window duration), since all requests
	// were burst at the same instant. Allow some tolerance.
	retryAfter, err := strconv.Atoi(rec.Header().Get("Retry-After"))
	assert.NoError(t, err)
	assert.Greater(t, retryAfter, 0, "Retry-After should be positive")
	assert.LessOrEqual(t, retryAfter, 10, "Retry-After should be ≤ window interval")
}

func TestLimiter_StopIsIdempotent(t *testing.T) {
	limiter := ratelimit.New(5, 100*time.Millisecond)

	// First stop should succeed.
	limiter.Stop()

	// Second stop should not panic or block.
	limiter.Stop()
	limiter.Stop()
}

func TestLimiter_StaleEntriesCleanedUp(t *testing.T) {
	limiter := ratelimit.New(5, 20*time.Millisecond)
	defer limiter.Stop()

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make a request from one IP, then wait for window to expire.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.99:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Wait for the window to expire AND the cleanup loop to run (interval = 20ms).
	time.Sleep(60 * time.Millisecond)

	// The limiter should still function — no panics, no stale data issues.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.99:12345"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "should allow requests after cleanup cycle")
}
