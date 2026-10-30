package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
)

// AdminAuthToken is the view of an AuthToken exposed to the administrators.
type AdminAuthToken struct {
	ID        bson.ObjectID    `json:"id"`
	Source    string           `json:"source"`
	Method    string           `json:"method"`
	Device    AdminDevice      `json:"device"`
	Location  *AdminIPLocation `json:"location,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	ExpiresAt time.Time        `json:"expires_at"`
}

// NewAdminAuthToken returns a new AdminAuthToken.
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

// HasLocation reports whether the token has a defined location.
func (token *AdminAuthToken) HasLocation() bool {
	return token.Location != nil && !token.Location.IsEmpty()
}

// URLParams returns the parameters to build the URL of the document: its ID.
func (token *AdminAuthToken) URLParams() map[string]string {
	return map[string]string{
		"id": token.ID.Hex(),
	}
}
