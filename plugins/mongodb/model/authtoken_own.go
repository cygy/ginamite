package model

import (
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
)

// OwnAuthToken is the view of an AuthToken exposed to its owner.
type OwnAuthToken struct {
	ID        bson.ObjectID  `json:"id"`
	Source    string         `json:"source"`
	Method    string         `json:"method"`
	Device    OwnDevice      `json:"device"`
	Location  *OwnIPLocation `json:"location,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// NewOwnAuthToken returns a new OwnAuthToken.
func NewOwnAuthToken(token *AuthToken) *OwnAuthToken {
	ownedToken := &OwnAuthToken{
		ID:        token.ID,
		Source:    token.Source,
		Method:    token.Method,
		Device:    NewOwnDevice(token.Device),
		CreatedAt: token.CreatedAt,
		ExpiresAt: token.ExpiresAt,
	}

	if token.Location != nil {
		ownedToken.Location = NewOwnIPLocation(*token.Location)
	}

	return ownedToken
}

// HasLocation reports whether the token has a defined location.
func (token *OwnAuthToken) HasLocation() bool {
	return token.Location != nil && !token.Location.IsEmpty()
}

// URLParams returns the parameters to build the URL of the document: its ID.
func (token *OwnAuthToken) URLParams() map[string]string {
	return map[string]string{
		"id": token.ID.Hex(),
	}
}
