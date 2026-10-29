package authtoken

import (
	"maps"
	"time"

	"github.com/cygy/ginamite/common/log"
)

// isAuthTokenPresent : returns true if the tokenID is found.
func (cache *authTokens) isAuthTokenPresent(token string) bool {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	_, ok := cache.tokens[token]
	return ok
}

// addAuthTokens : adds auth tokens to the list.
func (cache *authTokens) addAuthTokens(tokens map[string]time.Time) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	maps.Copy(cache.tokens, tokens)
}

// removeAuthToken : removes a token from the list.
func (cache *authTokens) removeAuthToken(token string) {
	cache.removeAuthTokens([]string{token})
}

// removeAuthTokens : removes tokens from the list.
func (cache *authTokens) removeAuthTokens(tokens []string) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	for _, token := range tokens {
		delete(cache.tokens, token)
	}
}

// getExpiredAuthTokens : get the expired tokens from the list.
func (cache *authTokens) getExpiredAuthTokens() []string {
	expiredTokens := []string{}
	now := time.Now()

	cache.mux.Lock()
	defer cache.mux.Unlock()

	for token, expirationDate := range cache.tokens {
		if expirationDate.Sub(now) < 0 {
			expiredTokens = append(expiredTokens, token)
		}
	}

	return expiredTokens
}

// deleteExpiredAuthTokens : deletes all the expired auth tokens every 'interval' seconds.
func (cache *authTokens) deleteExpiredAuthTokens(interval time.Duration) {
	for {
		<-time.After(interval * time.Second)
		go func(c *authTokens) {
			expiredTokens := c.getExpiredAuthTokens()
			if len(expiredTokens) > 0 {
				c.removeAuthTokens(expiredTokens)
				log.WithField("count", len(expiredTokens)).Info("remove expired tokens from the cache")
			}
		}(cache)
	}
}
