package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ── SecurityHeaders ──────────────────────────────────────────────────────────

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	headers := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
		"Referrer-Policy":           "strict-origin",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}
	for header, want := range headers {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
	if csp := rec.Header().Get("Content-Security-Policy"); csp == "" {
		t.Error("Content-Security-Policy should be set")
	}
}

// ── RateLimiter ──────────────────────────────────────────────────────────────

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(10, 5) // 10 req/s, burst 5
	ip := "127.0.0.1"

	// First 5 requests should be allowed (burst capacity).
	for i := 0; i < 5; i++ {
		if !rl.Allow(ip) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	// 6th request should be denied.
	if rl.Allow(ip) {
		t.Error("6th request should be denied (bucket exhausted)")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(100, 1) // 100 req/s, burst 1
	ip := "10.0.0.1"

	if !rl.Allow(ip) {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow(ip) {
		t.Error("second immediate request should be denied")
	}
	// Wait enough time for one token to refill (10ms at 100 req/s).
	time.Sleep(15 * time.Millisecond)
	if !rl.Allow(ip) {
		t.Error("request after refill should be allowed")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(10, 1) // burst 1
	if !rl.Allow("1.2.3.4") {
		t.Error("first request from 1.2.3.4 should be allowed")
	}
	if !rl.Allow("5.6.7.8") {
		t.Error("first request from 5.6.7.8 should be allowed (separate bucket)")
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	rl := NewRateLimiter(10, 1) // burst 1
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Limit(okHandler)

	makeReq := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := makeReq(); code != http.StatusOK {
		t.Errorf("first request code = %d, want 200", code)
	}
	if code := makeReq(); code != http.StatusTooManyRequests {
		t.Errorf("second request code = %d, want 429", code)
	}
}

// ── WSLimiter ────────────────────────────────────────────────────────────────

func TestWSLimiter_AcquireRelease(t *testing.T) {
	wl := NewWSLimiter(2)
	ip := "10.0.0.1"

	if !wl.Acquire(ip) {
		t.Fatal("first Acquire should succeed")
	}
	if !wl.Acquire(ip) {
		t.Fatal("second Acquire should succeed (max=2)")
	}
	if wl.Acquire(ip) {
		t.Error("third Acquire should fail (max=2)")
	}
	wl.Release(ip)
	if !wl.Acquire(ip) {
		t.Error("Acquire after Release should succeed")
	}
	wl.Release(ip)
	wl.Release(ip)
}

func TestWSLimiter_ReleaseUnderflow(t *testing.T) {
	// Release without Acquire should not panic or go negative.
	wl := NewWSLimiter(1)
	wl.Release("1.2.3.4") // should be a no-op
	if !wl.Acquire("1.2.3.4") {
		t.Error("Acquire after spurious Release should succeed")
	}
	wl.Release("1.2.3.4")
}

func TestWSLimiter_DifferentIPs(t *testing.T) {
	wl := NewWSLimiter(1)
	if !wl.Acquire("1.1.1.1") {
		t.Fatal("first IP should acquire")
	}
	if !wl.Acquire("2.2.2.2") {
		t.Fatal("second IP should acquire independently")
	}
	wl.Release("1.1.1.1")
	wl.Release("2.2.2.2")
}

// ── ClientIP ─────────────────────────────────────────────────────────────────

func TestClientIP(t *testing.T) {
	cases := []struct {
		remoteAddr string
		want       string
	}{
		{"192.168.1.1:54321", "192.168.1.1"},
		{"[::1]:8080", "::1"},
		{"10.0.0.1:1234", "10.0.0.1"},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = tc.remoteAddr
		got := ClientIP(req)
		if got != tc.want {
			t.Errorf("ClientIP(%q) = %q, want %q", tc.remoteAddr, got, tc.want)
		}
	}
}

func TestClientIP_NoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"
	got := ClientIP(req)
	if got != "192.168.1.1" {
		t.Errorf("ClientIP(no port) = %q, want 192.168.1.1", got)
	}
}

// ── RequireCSRF ──────────────────────────────────────────────────────────────

func TestRequireCSRF_GetPassthrough(t *testing.T) {
	called := false
	handler := RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		called = false
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if !called {
			t.Errorf("%s should pass through RequireCSRF", method)
		}
	}
}

func TestRequireCSRF_PostWithoutToken(t *testing.T) {
	handler := RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/something", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF token = %d, want 403", rec.Code)
	}
}

func TestRequireCSRF_DeleteWithoutToken(t *testing.T) {
	handler := RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/api/something", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("DELETE without CSRF token = %d, want 403", rec.Code)
	}
}
