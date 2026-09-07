package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimiter_AllowsRequestsWithinLimit(t *testing.T) {
	limiter := NewRateLimiter(
		rate.Limit(5),
		5,
	)

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/contact/", nil)
		req.RemoteAddr = "192.168.1.10:12345"

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"request %d: expected status %d, got %d",
				i+1,
				http.StatusOK,
				rec.Code,
			)
		}
	}
}

func TestRateLimiter_BlocksRequestsOverLimit(t *testing.T) {
	limiter := NewRateLimiter(
		rate.Limit(5),
		5,
	)

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/contact/", nil)
		req.RemoteAddr = "192.168.1.10:12345"

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/contact/", nil)
	req.RemoteAddr = "192.168.1.10:12345"

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTooManyRequests,
			rec.Code,
		)
	}
}

func TestRateLimiter_DifferentIPsHaveIndependentLimits(t *testing.T) {
	limiter := NewRateLimiter(
		rate.Limit(2),
		2,
	)

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/contact/", nil)
		req.RemoteAddr = "192.168.1.10:12345"

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("IP 1 request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/contact/", nil)
	req.RemoteAddr = "192.168.1.20:54321"

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"different IP: expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRateLimiter_AllowsRequestAfterRefill(t *testing.T) {
	limiter := NewRateLimiter(
		rate.Limit(10),
		1,
	)

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "192.168.1.10:12345"

	req := httptest.NewRequest(http.MethodPost, "/api/contact/", nil)
	req.RemoteAddr = ip

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/contact/", nil)
	req.RemoteAddr = ip

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected request to be rate limited, got %d", rec.Code)
	}

	time.Sleep(150 * time.Millisecond)

	req = httptest.NewRequest(http.MethodPost, "/api/contact/", nil)
	req.RemoteAddr = ip

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected request after refill to succeed, got %d",
			rec.Code,
		)
	}
}
