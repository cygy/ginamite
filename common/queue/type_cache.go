package queue

import "time"

// CacheAuthToken : content of a payload
type CacheAuthToken struct {
	Tokens map[string]time.Time `json:"tokens"`
}
