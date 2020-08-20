package cache

import "time"

// IsRevokedAuthToken : returns true if the tokenID is revoked.
func IsRevokedAuthToken(tokenID string) bool {
	return revokedAuthTokens.IsAuthTokenPresent(tokenID)
}

// AddRevokedAuthTokens : adds revoked auth tokens to the list.
func AddRevokedAuthTokens(tokens map[string]time.Time) {
	revokedAuthTokens.AddAuthTokens(tokens)
}

// RemoveRevokedAuthToken : removes a revoked auth token from the list.
func RemoveRevokedAuthToken(token string) {
	revokedAuthTokens.RemoveAuthToken(token)
}

// StartDeletingExpiredTokens : start deleting the expired auth tokens verry 'interval' seconds.
func StartDeletingExpiredTokens(interval time.Duration) {
	go revokedAuthTokens.DeleteExpiredAuthTokens(interval)
}
