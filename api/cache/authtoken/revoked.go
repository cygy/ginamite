package authtoken

import (
	"context"
	"time"
)

// Revoked is the set of the revoked auth tokens, kept in memory until they expire.
// It is safe for concurrent use. Create it with NewRevoked.
type Revoked struct {
	tokens *cache
}

// NewRevoked returns an empty set of revoked auth tokens.
func NewRevoked() *Revoked {
	return &Revoked{tokens: newCache()}
}

// IsRevoked reports whether the token ID is revoked.
func (r *Revoked) IsRevoked(tokenID string) bool {
	return r.tokens.has(tokenID)
}

// Add adds revoked tokens, with their expiration date, to the set.
func (r *Revoked) Add(tokens map[string]time.Time) {
	r.tokens.add(tokens)
}

// Remove removes a revoked token from the set.
func (r *Revoked) Remove(token string) {
	r.tokens.remove(token)
}

// StartDeletingExpired starts a background task that deletes the expired
// tokens every interval seconds, until ctx is done. The interval is a number of
// seconds, not a nanosecond-based duration, and must be positive.
func (r *Revoked) StartDeletingExpired(ctx context.Context, interval time.Duration) {
	go r.tokens.runDeleteExpired(ctx, interval*time.Second)
}
