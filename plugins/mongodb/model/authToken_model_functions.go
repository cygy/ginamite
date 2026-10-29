package model

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UpdateProperties : updates some properties of the token.
func (token *AuthToken) UpdateProperties(ctx context.Context, deviceName string, notificationsEnabled bool, notifications []string, db *mongo.Database) error {
	return token.Update(ctx, bson.M{"$set": bson.M{"device.name": deviceName, "device.notificationSettings.enabled": notificationsEnabled, "device.notificationSettings.notifications": notifications}}, db)
}

// UpdateExpirationDate : updates the expiration date of the token.
func (token *AuthToken) UpdateExpirationDate(ctx context.Context, expirationDate time.Time, db *mongo.Database) error {
	return token.Update(ctx, bson.M{"$set": bson.M{"expiresAt": expirationDate}}, db)
}

// UpdateLocation : updates some properties of the token.
func (token *AuthToken) UpdateLocation(ctx context.Context, location IPLocation, db *mongo.Database) error {
	updatedLocation := location
	updatedLocation.ID = bson.NewObjectID()

	return token.Update(ctx, bson.M{"$set": bson.M{"location": updatedLocation}}, db)
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
func (token *AuthToken) FromToken(t authentication.Token) error {
	userID, err := bson.ObjectIDFromHex(t.UserID)
	if err != nil {
		return errors.WrongObjectID()
	}
	token.UserID = userID
	token.Source = t.Source
	token.Key = t.PrivateKey
	token.CreatedAt = t.CreationDate
	token.ExpiresAt = t.ExpirationDate
	return nil
}
