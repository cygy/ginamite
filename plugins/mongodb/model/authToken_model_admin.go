package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// AdminAuthToken : admin version of the AuthToken struct
type AdminAuthToken struct {
	ID        bson.ObjectId    `json:"id"`
	Source    string           `json:"source"`
	Method    string           `json:"method"`
	Device    AdminDevice      `json:"device"`
	Location  *AdminIPLocation `json:"location,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	ExpiresAt time.Time        `json:"expires_at"`
}

// NewAdminAuthToken : returns a new 'AdminAuthToken' struct.
func NewAdminAuthToken(token *AuthToken) *AdminAuthToken {
	ownedToken := &AdminAuthToken{
		ID:        token.ID,
		Source:    token.Source,
		Method:    token.Method,
		Device:    NewAdminDevice(token.Device),
		CreatedAt: token.CreatedAt,
		ExpiresAt: token.ExpiresAt,
	}

	if token.Location != nil {
		ownedToken.Location = NewAdminIPLocation(*token.Location)
	}

	return ownedToken
}
