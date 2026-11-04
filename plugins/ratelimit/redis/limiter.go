package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/api/request"

	goredis "github.com/redis/go-redis/v9"
)

// script counts a request: it increments the counter of the key, gives it the period as expiration when it is created,
// and returns the count and the milliseconds left before the counter expires.
var script = goredis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return {count, redis.call("PTTL", KEYS[1])}
`)

// RateLimiter is a request.RateLimiter keeping its counters in Redis, shared by the instances of the API. It counts the
// requests of each client in fixed windows of the period, starting at the first request: a client can send up to twice
// the limit around the end of a window. The counters expire with their window.
type RateLimiter struct {
	client   goredis.Scripter
	prefix   string
	name     string
	requests int
	period   time.Duration
}

// NewRateLimiters returns the function creating the limiters of the groups of routes, to set as
// api.Server.Hooks.NewRateLimiter. The keys start with the prefix.
func NewRateLimiters(client goredis.Scripter, prefix string) request.NewRateLimiterFunc {
	return func(name string, requests int, period time.Duration) request.RateLimiter {
		return &RateLimiter{client: client, prefix: prefix, name: name, requests: requests, period: period}
	}
}

// Allow counts a request of the client and reports whether it is within the limit, and the delay to wait for
// otherwise, until the end of the window.
func (l *RateLimiter) Allow(ctx context.Context, client string) (ok bool, retryAfter time.Duration, err error) {
	key := fmt.Sprintf("%s%s:%s", l.prefix, l.name, client)

	result, err := script.Run(ctx, l.client, []string{key}, l.period.Milliseconds()).Int64Slice()
	if err != nil {
		return false, 0, fmt.Errorf("unable to count the request: %w", err)
	}
	if len(result) != 2 {
		return false, 0, fmt.Errorf("unable to count the request: unexpected reply %v", result)
	}

	if result[0] > int64(l.requests) {
		// A counter without expiration (PTTL < 0) can not happen with the script, but must never block a client forever.
		wait := time.Duration(result[1]) * time.Millisecond
		if result[1] < 0 {
			wait = l.period
		}

		return false, wait, nil
	}

	return true, 0, nil
}
