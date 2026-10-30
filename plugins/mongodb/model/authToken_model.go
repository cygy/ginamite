package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// AuthToken : authentication token properties.
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

// NewAuthToken : returns a new 'AuthToken' struct.
func NewAuthToken() *AuthToken {
	return &AuthToken{
		CreatedAt: time.Now(),
	}
}

// IsSaved : returns true if the 'AuthToken' document if saved in the collection.
func (token *AuthToken) IsSaved() bool {
	return bson.IsObjectIDProvided(token.ID)
}

// Save : inserts the 'AuthToken' document in the collection, returns an error if needed.
func (token *AuthToken) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, token, AuthTokenCollection, db)
}

// Update : updates the 'AuthToken' document, returns an error if needed.
func (token *AuthToken) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, token.ID, AuthTokenCollection, update, db)
}

// Delete : deletes the 'AuthToken' document from the collection, returns an error if needed.
func (token *AuthToken) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, token.ID, AuthTokenCollection, db)
}

// GetByID : initializes the 'AuthToken' struct from the 'AuthToken' document with the ID.
func (token *AuthToken) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, token, AuthTokenCollection, id, db)
}

// GetID : returns the 'AuthToken' ID.
func (token *AuthToken) GetID() string {
	return token.ID.Hex()
}

// SetID : initializes the 'AuthToken' ID.
func (token *AuthToken) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	token.ID = objectID
	return nil
}
