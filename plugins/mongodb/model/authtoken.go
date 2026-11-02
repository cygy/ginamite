package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"

	jwt "github.com/golang-jwt/jwt/v5"
)

// AuthTokenCollection is the name of the collection of the authToken documents.
const AuthTokenCollection = "authTokens"

// AuthToken is an authentication token issued to a user.
type AuthToken struct {
	ID        bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    bson.ObjectID `json:"-" bson:"userId,omitempty"`
	Source    string        `json:"source" bson:"source"`
	Method    string        `json:"method" bson:"method"`
	Device    Device        `json:"device" bson:"device"`
	Location  *IPLocation   `json:"location,omitempty" bson:"location,omitempty"`
	Key       string        `json:"key" bson:"key"`
	CreatedAt time.Time     `json:"created_at" bson:"createdAt"`
	ExpiresAt time.Time     `json:"expires_at" bson:"expiresAt"`
}

// NewAuthToken returns a new AuthToken.
func NewAuthToken() *AuthToken {
	return &AuthToken{
		CreatedAt: time.Now(),
	}
}

// IsSaved reports whether the AuthToken document is saved in the collection.
func (token *AuthToken) IsSaved() bool {
	return bson.IsObjectIDProvided(token.ID)
}

// Save inserts the AuthToken document in the collection, returning an error if needed.
func (token *AuthToken) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, token, AuthTokenCollection, db)
}

// Update updates the AuthToken document, returning an error if needed.
func (token *AuthToken) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, token.ID, AuthTokenCollection, update, db)
}

// Delete deletes the AuthToken document from the collection, returning an error if needed.
func (token *AuthToken) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, token.ID, AuthTokenCollection, db)
}

// GetByID loads the AuthToken document with the ID.
func (token *AuthToken) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, token, AuthTokenCollection, id, db)
}

// GetID returns the AuthToken ID.
func (token *AuthToken) GetID() string {
	return token.ID.Hex()
}

// SetID sets the ID of the AuthToken from its hexadecimal form. The ID is kept if the form is not valid.
func (token *AuthToken) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	token.ID = objectID
	return nil
}

// UpdateProperties updates some properties of the token.
func (token *AuthToken) UpdateProperties(ctx context.Context, deviceName string, notificationsEnabled bool, notifications []string, db store.Database) error {
	return token.Update(ctx, bson.M{"$set": bson.M{"device.name": deviceName, "device.notificationSettings.enabled": notificationsEnabled, "device.notificationSettings.notifications": notifications}}, db)
}

// UpdateExpirationDate updates the expiration date of the token.
func (token *AuthToken) UpdateExpirationDate(ctx context.Context, expirationDate time.Time, db store.Database) error {
	return token.Update(ctx, bson.M{"$set": bson.M{"expiresAt": expirationDate}}, db)
}

// UpdateLocation updates the location of the token.
func (token *AuthToken) UpdateLocation(ctx context.Context, location IPLocation, db store.Database) error {
	updatedLocation := location
	updatedLocation.ID = bson.NewObjectID()

	return token.Update(ctx, bson.M{"$set": bson.M{"location": updatedLocation}}, db)
}

// ToToken returns the authentication.Token of the auth token.
func (token *AuthToken) ToToken() authentication.Token {
	return authentication.Token{
		Source:         fmt.Sprintf("%s (%s)", token.Source, token.Method),
		UserID:         token.UserID.Hex(),
		PrivateKey:     token.Key,
		ExpirationDate: token.ExpiresAt,
		CreationDate:   token.CreatedAt,
	}
}

// FromToken fills the auth token from an authentication.Token. It returns an error if the user ID is not valid.
func (token *AuthToken) FromToken(t authentication.Token) error {
	userID, err := bson.ObjectIDFromHex(t.UserID)
	if err != nil {
		return err
	}
	token.UserID = userID
	token.Source = t.Source
	token.Key = t.PrivateKey
	token.CreatedAt = t.CreationDate
	token.ExpiresAt = t.ExpirationDate
	return nil
}

// GetAuthTokens returns all the auth tokens of a user.
func GetAuthTokens(ctx context.Context, userID string, db store.Database) ([]AuthToken, error) {
	tokens := []AuthToken{}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return tokens, err
	}

	err = document.GetDocumentsBySelectorAndSort(ctx, &tokens, AuthTokenCollection, bson.M{"userId": objectID}, []string{"userId", "-createdAt"}, db)
	return tokens, err
}

// UserFromAuthToken returns a User with the values of the claims of a JWT: the ID, the nickname, the locale and the timezone.
func UserFromAuthToken(token *jwt.Token) (*User, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("the claims of the token are not a map")
	}

	var userID bson.ObjectID
	if sub, ok := claims["sub"].(string); ok {
		objectID, err := bson.ObjectIDFromHex(sub)
		if err != nil {
			return nil, err
		}
		userID = objectID
	}

	nickname, _ := claims["nickname"].(string)
	locale, _ := claims["locale"].(string)
	timezone, _ := claims["timezone"].(string)

	return &User{
		ID:       userID,
		Username: nickname,
		Settings: UserSettings{
			Locale:   locale,
			Timezone: timezone,
		},
	}, nil
}

// DeleteAuthTokenByID deletes the document with the ID.
func DeleteAuthTokenByID(ctx context.Context, tokenID string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(tokenID)
	if err != nil {
		return err
	}
	return document.DeleteDocument(ctx, objectID, AuthTokenCollection, db)
}

// DeleteAuthTokensByUser deletes the documents from the user.
func DeleteAuthTokensByUser(ctx context.Context, userID string, db store.Database) (int64, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}
	return document.DeleteDocumentsBySelector(ctx, AuthTokenCollection, bson.M{"userId": objectID}, db)
}

// DeleteExpiredAuthTokens deletes the expired auth tokens.
func DeleteExpiredAuthTokens(ctx context.Context, db store.Database) (int64, error) {
	return document.DeleteDocumentsBySelector(ctx, AuthTokenCollection, bson.M{"expiresAt": bson.M{"$lt": time.Now()}}, db)
}

// IsValidAuthToken checks that the auth token is valid (non-expired and non-revoked).
func IsValidAuthToken(ctx context.Context, tokenID string, db store.Database) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(tokenID)
	if err != nil {
		return false, err
	}
	count, err := document.CountBySelector(ctx, AuthTokenCollection, bson.D{
		{Key: "_id", Value: objectID},
		{Key: "expiresAt", Value: bson.M{"$gt": time.Now()}},
	}, db)
	return count == 1, err
}
