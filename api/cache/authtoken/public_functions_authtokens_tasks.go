package authtoken

import "time"

// StartDeletingExpiredTokens : start deleting the expired auth tokens verry 'interval' seconds.
func StartDeletingExpiredTokens(interval time.Duration) {
	go revokedAuthTokens.deleteExpiredAuthTokens(interval)
}
