package queue

import "time"

// InvalidateCache : invalidates a cache.
func InvalidateCache(message string, payload interface{}) {
	sendMessage(message, payload, TopicCache)
}

// InvalidAuthToken : invalidates a auth token.
func InvalidAuthToken(tokenID string, expirationDate time.Time) {
	payload := CacheAuthToken{
		Tokens: map[string]time.Time{
			tokenID: expirationDate,
		},
	}

	InvalidateCache(MessageCacheInvalidAuthToken, payload)
}

// InvalidAuthTokens : invalidates auth tokens.
func InvalidAuthTokens(tokens map[string]time.Time) {
	payload := CacheAuthToken{
		Tokens: tokens,
	}

	InvalidateCache(MessageCacheInvalidAuthToken, payload)
}
