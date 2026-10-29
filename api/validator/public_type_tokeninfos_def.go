package validator

import "time"

// TokenInfos struct about the informations of a token
type TokenInfos struct {
	Token     string
	UserID    string
	Email     string
	LastName  string
	FirstName string
	CreatedAt time.Time
	ExpiresAt time.Time
}
