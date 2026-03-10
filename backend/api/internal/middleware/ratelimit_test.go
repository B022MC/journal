package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeLimiter struct {
	allow bool
	err   error
	hits  int
}

func (f *fakeLimiter) Allow(context.Context, string) (bool, int, error) {
	f.hits++
	return f.allow, 0, f.err
}

func (f *fakeLimiter) Window() time.Duration {
	return time.Minute
}

func TestRateLimitBlocksSearch(t *testing.T) {
	searchLimiter := &fakeLimiter{allow: false}
	mw := &RateLimitMiddleware{
		rules: []routeRule{
			{
				method:  http.MethodGet,
				match:   matchSearchPath,
				keyFunc: requestIP,
				limiter: searchLimiter,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/search?query=test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()

	called := false
	mw.Handle(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})(rr, req)

	if called {
		t.Fatal("expected request to be blocked by rate limiter")
	}
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") != "60" {
		t.Fatalf("expected Retry-After=60, got %q", rr.Header().Get("Retry-After"))
	}
}

func TestRateLimitAllowsWhenLimiterErrors(t *testing.T) {
	searchLimiter := &fakeLimiter{allow: true, err: errors.New("redis down")}
	mw := &RateLimitMiddleware{
		rules: []routeRule{
			{
				method:  http.MethodGet,
				match:   matchSearchPath,
				keyFunc: requestIP,
				limiter: searchLimiter,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/search?query=test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()

	called := false
	mw.Handle(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})(rr, req)

	if !called {
		t.Fatal("expected request to pass through on limiter error")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestRequestUserKeyPrefersJwtUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/papers/1/rate", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	ctx := context.WithValue(req.Context(), "userId", json.Number("42"))
	req = req.WithContext(ctx)

	if got := requestUserKey(req); got != "42" {
		t.Fatalf("expected user key 42, got %q", got)
	}
}

func TestRequestIPStripsForwardedPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/search?query=test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.8:3456, 10.0.0.9")

	if got := requestIP(req); got != "10.0.0.8" {
		t.Fatalf("expected forwarded ip without port, got %q", got)
	}
}

func TestPathMatchers(t *testing.T) {
	cases := []struct {
		name string
		fn   func(string) bool
		path string
		want bool
	}{
		{name: "search", fn: matchSearchPath, path: "/api/v1/papers/search", want: true},
		{name: "paper-rate", fn: matchPaperRatePath, path: "/api/v1/papers/12/rate", want: true},
		{name: "paper-flag", fn: matchFlagPath, path: "/api/v1/papers/12/flag", want: true},
		{name: "rating-flag", fn: matchFlagPath, path: "/api/v1/ratings/12/flag", want: true},
		{name: "bad-paper-rate", fn: matchPaperRatePath, path: "/api/v1/papers/12/rate/again", want: false},
	}

	for _, tc := range cases {
		if got := tc.fn(tc.path); got != tc.want {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, got)
		}
	}
}
