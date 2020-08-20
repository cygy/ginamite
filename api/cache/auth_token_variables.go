package cache

import "time"

var revokedAuthTokens = AuthTokens{
	tokens: make(map[string]time.Time),
}
