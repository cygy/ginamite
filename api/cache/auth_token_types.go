package cache

import (
	"sync"
	"time"
)

// AuthTokens : cached auth tokens.
type AuthTokens struct {
	mux    sync.Mutex
	tokens map[string]time.Time
}
