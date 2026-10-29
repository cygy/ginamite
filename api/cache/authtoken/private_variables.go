package authtoken

import "time"

var revokedAuthTokens = authTokens{
	tokens: make(map[string]time.Time),
}
