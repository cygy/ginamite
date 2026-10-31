package request

import (
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
	if ok, _ := l.Allow("client"); !ok || l.Size() != 0 {
		t.Error("a nil limiter allows everything")
	}
	if RateLimit(nil) != nil || RateLimitFromConfiguration(config.RateLimitConfiguration{}) != nil {
		t.Error("no middleware without a limit")
	}
}

func TestLimiter_Allow(t *testing.T) {
	l, advance := clockedLimiter()

	for i := range 3 {
		if ok, _ := l.Allow("a"); !ok {
			t.Fatalf("request %d of the burst refused", i+1)
		}
	}

	ok, retryAfter := l.Allow("a")
	if ok || retryAfter != 10*time.Second {
		t.Errorf("4th request: ok = %v, retry after %v, want refused and 10s", ok, retryAfter)
	}

	// Another client has its own bucket.
	if ok, _ := l.Allow("b"); !ok {
		t.Error("a client must not be limited by another one")
	}

	// A token comes back every 10 seconds.
	advance(10 * time.Second)
	if ok, _ := l.Allow("a"); !ok {
		t.Error("a token must be back after 10s")
	}
	if ok, _ := l.Allow("a"); ok {
		t.Error("only one token came back")
	}

	// The bucket never holds more than the burst.
	advance(time.Hour)
	for range 3 {
		l.Allow("a")
	}
	if ok, _ := l.Allow("a"); ok {
		t.Error("the burst must stay capped")
	}
}

func TestLimiter_Sweep(t *testing.T) {
	l, advance := clockedLimiter()

	l.Allow("a")
	l.Allow("b")
	if l.Size() != 2 {
		t.Fatalf("size = %d", l.Size())
	}

	advance(time.Minute)
	l.Allow("c")
	if l.Size() != 1 {
		t.Errorf("size = %d, want the idle clients forgotten", l.Size())
	}
}

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimitFromConfiguration(config.RateLimitConfiguration{Requests: 2, Seconds: 60}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	serve := func(remote string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
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
