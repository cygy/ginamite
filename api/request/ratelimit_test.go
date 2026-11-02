package request

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

// clockedLimiter returns a limiter of 3 requests per 30 seconds, and a function that moves its clock.
func clockedLimiter() (*Limiter, func(time.Duration)) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l := NewLimiter(3, 30*time.Second)
	l.now = func() time.Time { return now }

	return l, func(d time.Duration) { now = now.Add(d) }
}

func TestNewLimiter_Disabled(t *testing.T) {
	for _, l := range []*Limiter{NewLimiter(0, time.Minute), NewLimiter(5, 0), NewLimiter(-1, time.Minute)} {
		if l != nil {
			t.Error("a limiter without limit must be nil")
		}
	}

	var l *Limiter
	if ok, _, err := l.Allow(t.Context(), "client"); !ok || err != nil || l.Size() != 0 {
		t.Error("a nil limiter allows everything")
	}
	if RateLimit(nil) != nil || RateLimitFromConfiguration("a", config.RateLimitConfiguration{}, nil) != nil {
		t.Error("no middleware without a limit")
	}
}

func TestLimiter_Allow(t *testing.T) {
	l, advance := clockedLimiter()

	for i := range 3 {
		if ok, _, _ := l.Allow(t.Context(), "a"); !ok {
			t.Fatalf("request %d of the burst refused", i+1)
		}
	}

	ok, retryAfter, _ := l.Allow(t.Context(), "a")
	if ok || retryAfter != 10*time.Second {
		t.Errorf("4th request: ok = %v, retry after %v, want refused and 10s", ok, retryAfter)
	}

	// Another client has its own bucket.
	if ok, _, _ := l.Allow(t.Context(), "b"); !ok {
		t.Error("a client must not be limited by another one")
	}

	// A token comes back every 10 seconds.
	advance(10 * time.Second)
	if ok, _, _ := l.Allow(t.Context(), "a"); !ok {
		t.Error("a token must be back after 10s")
	}
	if ok, _, _ := l.Allow(t.Context(), "a"); ok {
		t.Error("only one token came back")
	}

	// The bucket never holds more than the burst.
	advance(time.Hour)
	for range 3 {
		_, _, _ = l.Allow(t.Context(), "a")
	}
	if ok, _, _ := l.Allow(t.Context(), "a"); ok {
		t.Error("the burst must stay capped")
	}
}

func TestLimiter_Sweep(t *testing.T) {
	l, advance := clockedLimiter()

	_, _, _ = l.Allow(t.Context(), "a")
	_, _, _ = l.Allow(t.Context(), "b")
	if l.Size() != 2 {
		t.Fatalf("size = %d", l.Size())
	}

	advance(time.Minute)
	_, _, _ = l.Allow(t.Context(), "c")
	if l.Size() != 1 {
		t.Errorf("size = %d, want the idle clients forgotten", l.Size())
	}
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimitFromConfiguration("a", config.RateLimitConfiguration{Requests: 2, Seconds: 60}, nil))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	serve := func(remote string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		req.RemoteAddr = remote + ":1234"
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		return recorder
	}

	for range 2 {
		if code := serve("192.0.2.1").Code; code != http.StatusOK {
			t.Fatalf("status = %d", code)
		}
	}

	limited := serve("192.0.2.1")
	if limited.Code != http.StatusTooManyRequests || limited.Header().Get("Retry-After") != "30" {
		t.Errorf("status = %d, Retry-After = %q, want 429 and 30", limited.Code, limited.Header().Get("Retry-After"))
	}

	if code := serve("192.0.2.2").Code; code != http.StatusOK {
		t.Errorf("another client: status = %d", code)
	}
}

// fakeLimiter is a RateLimiter answering what the test sets.
type fakeLimiter struct {
	ok     bool
	err    error
	client string
}

func (f *fakeLimiter) Allow(_ context.Context, client string) (bool, time.Duration, error) {
	f.client = client

	return f.ok, time.Minute, f.err
}

func TestRateLimitFromConfiguration_Hook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		limiter *fakeLimiter
		want    int
		errors  int
	}{
		{"allowed", &fakeLimiter{ok: true}, http.StatusOK, 0},
		{"refused", &fakeLimiter{}, http.StatusTooManyRequests, 0},
		{"failure lets the request through", &fakeLimiter{err: errors.New("boom")}, http.StatusOK, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotName string
			var gotRequests int
			var gotPeriod time.Duration
			middleware := RateLimitFromConfiguration("authentication", config.RateLimitConfiguration{Requests: 2, Seconds: 60},
				func(name string, requests int, period time.Duration) RateLimiter {
					gotName, gotRequests, gotPeriod = name, requests, period
					return tt.limiter
				})
			if gotName != "authentication" || gotRequests != 2 || gotPeriod != time.Minute {
				t.Errorf("limiter created with %q, %d, %v", gotName, gotRequests, gotPeriod)
			}

			var errs int
			router := gin.New()
			router.Use(middleware)
			router.GET("/", func(c *gin.Context) { errs = len(c.Errors); c.Status(http.StatusOK) })

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			req.RemoteAddr = "192.0.2.1:1234"
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.want || errs != tt.errors || tt.limiter.client != "192.0.2.1" {
				t.Errorf("status = %d, errors = %d, client = %q, want %d, %d errors and the IP", recorder.Code, errs, tt.limiter.client, tt.want, tt.errors)
			}
		})
	}
}
