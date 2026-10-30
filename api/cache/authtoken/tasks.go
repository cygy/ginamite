package authtoken

import "time"

// StartDeletingExpired starts a background task that deletes the expired
// tokens every interval seconds. The interval is a number of seconds, not a
// nanosecond-based duration, and must be positive.
func StartDeletingExpired(interval time.Duration) {
	go revoked.runDeleteExpired(interval * time.Second)
}
