package cache

import (
	"time"

	"github.com/cygy/ginamite/common/log"
)

// IsAuthTokenPresent : returns true if the tokenID is found.
func (cache *AuthTokens) IsAuthTokenPresent(token string) bool {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	_, ok := cache.tokens[token]
	return ok
}

// AddAuthTokens : adds auth tokens to the list.
func (cache *AuthTokens) AddAuthTokens(tokens map[string]time.Time) {
	cache.mux.Lock()

	for token, expirationDate := range tokens {
		cache.tokens[token] = expirationDate
	}

	cache.mux.Unlock()
}

// RemoveAuthToken : removes a token from the list.
func (cache *AuthTokens) RemoveAuthToken(token string) {
	cache.RemoveAuthTokens([]string{token})
}

// RemoveAuthTokens : removes tokens from the list.
func (cache *AuthTokens) RemoveAuthTokens(tokens []string) {
	cache.mux.Lock()

	for _, token := range tokens {
		if _, ok := cache.tokens[token]; ok {
			delete(cache.tokens, token)
		}
	}

	cache.mux.Unlock()
}

// GetExpiredAuthTokens : get the expired tokens from the list.
func (cache *AuthTokens) GetExpiredAuthTokens() []string {
	expiredTokens := []string{}
	now := time.Now()

	cache.mux.Lock()

	for token, expirationDate := range cache.tokens {
		if expirationDate.Sub(now) < 0 {
			expiredTokens = append(expiredTokens, token)
		}
	}

	cache.mux.Unlock()

	return expiredTokens
}

// DeleteExpiredAuthTokens : deletes all the expired auth tokens every 'interval' seconds.
func (cache *AuthTokens) DeleteExpiredAuthTokens(interval time.Duration) {
	for {
		<-time.After(interval * time.Second)
		go func(c *AuthTokens) {
			expiredTokens := c.GetExpiredAuthTokens()
			if len(expiredTokens) > 0 {
				c.RemoveAuthTokens(expiredTokens)
				log.WithField("count", len(expiredTokens)).Info("remove expired tokens from the cache")
			}
		}(cache)
	}
}
