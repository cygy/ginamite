package request

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

// RateLimiter decides whether a client can send a request now. The API uses a Limiter, kept in memory, by default: with
// several instances of the API each one counts the requests it receives, so a client can send as many requests as the
// limit times the number of instances. Set api.Server.Hooks.NewRateLimiter to share the counters, in a database for
// instance.
type RateLimiter interface {
	// Allow reports whether the client can send a request now, and the delay to wait for otherwise. An error lets the
	// request through: the limit must not stop the API when its storage fails.
	Allow(ctx context.Context, client string) (ok bool, retryAfter time.Duration, err error)
}

// NewRateLimiterFunc returns the RateLimiter allowing requests in period to each client, for the group of routes with
// the name, such as "authentication": the groups have their own counters.
type NewRateLimiterFunc func(name string, requests int, period time.Duration) RateLimiter

// Limiter is a RateLimiter keeping its counters in memory, by instance of the server. It limits the requests of each
// client to a number in a period with a token bucket: a client can send a burst of requests, then as many as the period
// allows.
type Limiter struct {
	mu        sync.Mutex
	requests  float64
	period    time.Duration
	buckets   map[string]*bucket
	lastSweep time.Time
	now       func() time.Time
}

// bucket is the tokens of a client.
type bucket struct {
	tokens float64
	last   time.Time
}

// NewLimiter returns a Limiter allowing requests in period to each client. It returns nil, which allows everything,
// if requests or period is not positive.
func NewLimiter(requests int, period time.Duration) *Limiter {
	if requests <= 0 || period <= 0 {
		return nil
	}

	return &Limiter{requests: float64(requests), period: period, buckets: map[string]*bucket{}, now: time.Now}
}

// Allow reports whether the client can send a request now, and the delay to wait for otherwise. It never fails.
func (l *Limiter) Allow(_ context.Context, client string) (ok bool, retryAfter time.Duration, err error) {
	if l == nil {
		return true, 0, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	b, found := l.buckets[client]
	if !found {
		b = &bucket{tokens: l.requests, last: now}
		l.buckets[client] = b
	}

	// The bucket fills at the rate of the limit, up to the burst.
	rate := l.requests / l.period.Seconds()
	b.tokens = min(l.requests, b.tokens+now.Sub(b.last).Seconds()*rate)
	b.last = now

	if b.tokens >= 1 {
		b.tokens--

		return true, 0, nil
	}

	return false, time.Duration((1 - b.tokens) / rate * float64(time.Second)), nil
}

// sweep forgets the clients whose bucket is full again, at most once per period: it keeps the memory bounded.
func (l *Limiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.period {
		return
	}
	l.lastSweep = now

	for client, b := range l.buckets {
		if now.Sub(b.last) >= l.period {
			delete(l.buckets, client)
		}
	}
}

// Size returns the number of clients the limiter remembers.
func (l *Limiter) Size() int {
	if l == nil {
		return 0
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.buckets)
}

// RateLimit returns the middleware that answers 429 Too Many Requests, with the header Retry-After, to a client who sent
// too many requests. The client is identified by its IP address: see the setting server.trusted_proxies. A failure of
// the limiter is saved in the errors of the context and lets the request through. It returns nil, which means no
// limit, if the limiter is nil.
func RateLimit(l RateLimiter) gin.HandlerFunc {
	if l == nil {
		return nil
	}

	return func(c *gin.Context) {
		ok, retryAfter, err := l.Allow(c, c.ClientIP())
		if err != nil {
			c.Error(fmt.Errorf("unable to limit the requests: %w", err))
			return
		}
		if !ok {
			response.TooManyRequests(c, retryAfter)
		}
	}
}

// RateLimitFromConfiguration returns the middleware limiting the requests of the group of routes with the name as the
// configuration says, or nil if it does not limit them. The limiter is the one returned by newLimiter, or a Limiter
// if newLimiter is nil.
func RateLimitFromConfiguration(name string, c config.RateLimitConfiguration, newLimiter NewRateLimiterFunc) gin.HandlerFunc {
	if c.Requests <= 0 || c.Seconds <= 0 {
		return nil
	}

	period := time.Duration(c.Seconds) * time.Second
	if newLimiter == nil {
		return RateLimit(NewLimiter(c.Requests, period))
	}

	return RateLimit(newLimiter(name, c.Requests, period))
}
