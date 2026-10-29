package authtoken

import "time"

// IsRevokedAuthToken : returns true if the tokenID is revoked.
func IsRevokedAuthToken(tokenID string) bool {
	return revokedAuthTokens.isAuthTokenPresent(tokenID)
}

// AddRevokedAuthTokens : adds revoked auth tokens to the list.
func AddRevokedAuthTokens(tokens map[string]time.Time) {
	revokedAuthTokens.addAuthTokens(tokens)
}

// RemoveRevokedAuthToken : removes a revoked auth token from the list.
func RemoveRevokedAuthToken(token string) {
	revokedAuthTokens.removeAuthToken(token)
}
