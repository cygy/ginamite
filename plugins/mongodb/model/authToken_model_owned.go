package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// OwnAuthToken : public version of the AuthToken struct
type OwnAuthToken struct {
	ID        bson.ObjectId  `json:"id"`
	Source    string         `json:"source"`
	Method    string         `json:"method"`
	Device    OwnDevice      `json:"device"`
	Location  *OwnIPLocation `json:"location,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// NewOwnAuthToken : returns a new 'OwnAuthToken' struct.
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
