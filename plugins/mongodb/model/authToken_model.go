package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// AuthToken : authentication token properties.
type AuthToken struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	UserID    bson.ObjectId `json:"-" bson:"userId,omitempty"`
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
	return token.ID.Valid()
}

// Save : inserts the 'AuthToken' document in the collection, returns an error if needed.
func (token *AuthToken) Save(db *mgo.Database) (string, error) {
	return document.SaveDocument(token, AuthTokenCollection, db)
}

// Update : updates the 'AuthToken' document, returns an error if needed.
func (token *AuthToken) Update(update interface{}, db *mgo.Database) error {
	return document.UpdateDocument(token.ID, AuthTokenCollection, update, db)
}

// Delete : deletes the 'AuthToken' document from the collection, returns an error if needed.
func (token *AuthToken) Delete(db *mgo.Database) error {
	return document.DeleteDocument(token.ID, AuthTokenCollection, db)
}

// GetByID : initializes the 'AuthToken' struct from the 'AuthToken' document with the ID.
func (token *AuthToken) GetByID(id string, db *mgo.Database) error {
	return document.GetDocumentByID(token, AuthTokenCollection, id, db)
}

// GetID : returns the 'AuthToken' ID.
func (token *AuthToken) GetID() string {
	return token.ID.Hex()
}

// SetID : initializes the 'AuthToken' ID.
func (token *AuthToken) SetID(id string) {
	token.ID = bson.ObjectIdHex(id)
}
