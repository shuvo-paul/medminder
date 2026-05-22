package ratelimit

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Limiter implements IP-based rate limiting with a sliding window.
// Thread-safe for concurrent use.
type Limiter struct {
	rate     int
	interval time.Duration
	mu       sync.Mutex
	windows  map[string][]time.Time
	now      func() time.Time
	done     chan struct{}
}

// New creates a new Limiter and starts a background goroutine that
// periodically removes stale IP entries to prevent memory leaks.
//
// rate: max requests per interval per IP
// interval: sliding window duration
func New(rate int, interval time.Duration) *Limiter {
	l := &Limiter{
		rate:     rate,
		interval: interval,
		windows:  make(map[string][]time.Time),
		now:      time.Now,
		done:     make(chan struct{}),
	}
	go l.cleanupLoop()
	return l
}

// Stop stops the limiter's background cleanup goroutine.
// Safe to call multiple times; subsequent calls are no-ops.
// After Stop returns, the limiter still functions but stale entries
// are only removed on access, not proactively scanned.
func (l *Limiter) Stop() {
	select {
	case <-l.done:
		// Already stopped.
	default:
		close(l.done)
	}
}

// cleanupLoop periodically removes IP entries whose windows have no
// active requests. Runs once per interval.
func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	for {
		select {
		case <-l.done:
			return
		case <-ticker.C:
			l.pruneEmpty()
		}
	}
}

// pruneEmpty removes IP keys with empty time windows.
func (l *Limiter) pruneEmpty() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, window := range l.windows {
		if len(window) == 0 {
			delete(l.windows, ip)
		}
	}
}

// Middleware returns a Chi-compatible HTTP middleware that rate-limits by IP.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		allowed, remaining, resetAt := l.allow(ip)

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.rate))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(int(time.Until(resetAt).Seconds())))
			http.Error(w, "429 too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow checks and records a request from the given IP.
// Returns (allowed bool, remaining int, resetAt time.Time).
func (l *Limiter) allow(ip string) (bool, int, time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.interval)
	window := l.windows[ip]

	// Prune entries outside the window.
	var valid []time.Time
	for _, t := range window {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	remaining := l.rate - len(valid)

	// Determine the reset time: when the oldest active entry expires.
	// If no entries, reset is now + interval.
	var resetAt time.Time
	if len(valid) > 0 {
		resetAt = valid[0].Add(l.interval)
	} else {
		resetAt = now.Add(l.interval)
	}

	if remaining <= 0 {
		// Save pruned window so next check doesn't re-scan stale entries.
		l.windows[ip] = valid
		return false, 0, resetAt
	}

	l.windows[ip] = append(valid, now)
	return true, remaining - 1, resetAt
}
