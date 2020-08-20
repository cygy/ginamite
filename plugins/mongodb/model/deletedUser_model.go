package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// DeletedUser : deleted user properties.
type DeletedUser struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	UserID    bson.ObjectId `json:"-" bson:"userId,omitempty"`
	CreatedAt time.Time     `json:"created_at" bson:"createdAt"`
}

// NewDeletedUser : returns a new 'DeletedUser' struct.
func NewDeletedUser(userID string) *DeletedUser {
	return &DeletedUser{
		UserID:    bson.ObjectIdHex(userID),
		CreatedAt: time.Now(),
	}
}

// IsSaved : returns true if the 'DeletedUser' document if saved in the collection.
func (deletedUser *DeletedUser) IsSaved() bool {
	return deletedUser.ID.Valid()
}

// Save : inserts the 'DeletedUser' document in the collection, returns an error if needed.
func (deletedUser *DeletedUser) Save(db *mgo.Database) (string, error) {
	return document.SaveDocument(deletedUser, DeletedUserCollection, db)
}

// Update : updates the 'DeletedUser' document, returns an error if needed.
func (deletedUser *DeletedUser) Update(update interface{}, db *mgo.Database) error {
	return document.UpdateDocument(deletedUser.ID, DeletedUserCollection, update, db)
}

// Delete : deletes the 'DeletedUser' document from the collection, returns an error if needed.
func (deletedUser *DeletedUser) Delete(db *mgo.Database) error {
	return document.DeleteDocument(deletedUser.ID, DeletedUserCollection, db)
}

// GetByID : initializes the 'DeletedUser' struct from the 'DeletedUser' document with the ID.
func (deletedUser *DeletedUser) GetByID(id string, db *mgo.Database) error {
	return document.GetDocumentByID(deletedUser, DeletedUserCollection, id, db)
}

// GetID : returns the 'DeletedUser' ID.
func (deletedUser *DeletedUser) GetID() string {
	return deletedUser.ID.Hex()
}

// SetID : initializes the 'DeletedUser' ID.
func (deletedUser *DeletedUser) SetID(id string) {
	deletedUser.ID = bson.ObjectIdHex(id)
}
