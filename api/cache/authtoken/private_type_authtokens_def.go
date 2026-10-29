package authtoken

import (
	"sync"
	"time"
)

// authTokens : cached auth tokens.
type authTokens struct {
	mux    sync.Mutex
	tokens map[string]time.Time
}
