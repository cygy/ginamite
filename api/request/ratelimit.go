package request

import (
	"sync"
	"time"

	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

// Limiter limits the requests of each client to a number in a period, with a token bucket: a client can send a burst
// of requests, then as many as the period allows. The counters are kept in memory, by instance of the server.
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

// Allow reports whether the client can send a request now, and the delay to wait for otherwise.
func (l *Limiter) Allow(client string) (ok bool, retryAfter time.Duration) {
	if l == nil {
		return true, 0
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

		return true, 0
	}

	return false, time.Duration((1 - b.tokens) / rate * float64(time.Second))
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
// too many requests. The client is identified by its IP address: see the setting server.trusted_proxies. It returns nil,
// which means no limit, if the limiter is nil.
func RateLimit(l *Limiter) gin.HandlerFunc {
	if l == nil {
		return nil
	}

	return func(c *gin.Context) {
		if ok, retryAfter := l.Allow(c.ClientIP()); !ok {
			response.TooManyRequests(c, retryAfter)
		}
	}
}

// RateLimitFromConfiguration returns the middleware limiting the requests as the configuration says, or nil if it
// does not limit them.
func RateLimitFromConfiguration(c config.RateLimitConfiguration) gin.HandlerFunc {
	return RateLimit(NewLimiter(c.Requests, time.Duration(c.Seconds)*time.Second))
}
