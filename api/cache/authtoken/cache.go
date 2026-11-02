// Package authtoken caches the revoked auth tokens in memory.
package authtoken

import (
	"context"
	"maps"
	"sync"
	"time"

	"github.com/cygy/ginamite/common/log"
)

// cache is a concurrency-safe set of tokens with their expiration date.
type cache struct {
	mux    sync.Mutex
	tokens map[string]time.Time
}

// newCache returns an empty cache ready to use.
func newCache() *cache {
	return &cache{tokens: make(map[string]time.Time)}
}

// has reports whether the token is in the cache.
func (c *cache) has(token string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()

	_, ok := c.tokens[token]
	return ok
}

// add stores the tokens with their expiration date.
func (c *cache) add(tokens map[string]time.Time) {
	c.mux.Lock()
	defer c.mux.Unlock()

	maps.Copy(c.tokens, tokens)
}

// remove deletes the tokens from the cache.
func (c *cache) remove(tokens ...string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	for _, token := range tokens {
		delete(c.tokens, token)
	}
}

// deleteExpired removes the tokens expired at now and returns how many were removed.
func (c *cache) deleteExpired(now time.Time) int {
	c.mux.Lock()
	defer c.mux.Unlock()

	count := 0
	for token, expirationDate := range c.tokens {
		if expirationDate.Before(now) {
			delete(c.tokens, token)
			count++
		}
	}

	return count
}

// runDeleteExpired deletes the expired tokens every interval, until ctx is done.
func (c *cache) runDeleteExpired(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if count := c.deleteExpired(now); count > 0 {
				log.Info("remove expired tokens from the cache", "count", count)
			}
		}
	}
}
