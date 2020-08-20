package authentication

import "time"

// Token : The content of an auth token.
type Token struct {
	UserID         string
	Source         string
	PrivateKey     string
	ExpirationDate time.Time
	CreationDate   time.Time
	Other          map[string]interface{}
}
