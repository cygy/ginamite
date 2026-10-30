package authtoken

import "time"

var revoked = newCache()

// IsRevoked reports whether the token ID is revoked.
func IsRevoked(tokenID string) bool {
	return revoked.has(tokenID)
}

// Add adds revoked tokens, with their expiration date, to the cache.
func Add(tokens map[string]time.Time) {
	revoked.add(tokens)
}

// Remove removes a revoked token from the cache.
func Remove(token string) {
	revoked.remove(token)
}
