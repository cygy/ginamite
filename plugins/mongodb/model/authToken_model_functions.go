package model

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UpdateProperties : updates some properties of the token.
func (token *AuthToken) UpdateProperties(deviceName string, notificationsEnabled bool, notifications []string, db *mgo.Database) error {
	return token.Update(bson.M{"$set": bson.M{"device.name": deviceName, "device.notificationSettings.enabled": notificationsEnabled, "device.notificationSettings.notifications": notifications}}, db)
}

// UpdateExpirationDate : updates the expiration date of the token.
func (token *AuthToken) UpdateExpirationDate(expirationDate time.Time, db *mgo.Database) error {
	return token.Update(bson.M{"$set": bson.M{"expiresAt": expirationDate}}, db)
}

// UpdateLocation : updates some properties of the token.
func (token *AuthToken) UpdateLocation(location IPLocation, db *mgo.Database) error {
	updatedLocation := location
	updatedLocation.ID = ""

	return token.Update(bson.M{"$set": bson.M{"location": updatedLocation}}, db)
}

// ToToken : returns an 'authentication.Token' struct from this struct.
func (token *AuthToken) ToToken() authentication.Token {
	return authentication.Token{
		Source:         fmt.Sprintf("%s (%s)", token.Source, token.Method),
		UserID:         token.UserID.Hex(),
		PrivateKey:     token.Key,
		ExpirationDate: token.ExpiresAt,
		CreationDate:   token.CreatedAt,
	}
}

// FromToken : fills this struct from an 'authentication.Token'.
func (token *AuthToken) FromToken(t authentication.Token) {
	token.UserID = bson.ObjectIdHex(t.UserID)
	token.Source = t.Source
	token.Key = t.PrivateKey
	token.CreatedAt = t.CreationDate
	token.ExpiresAt = t.ExpirationDate
}
