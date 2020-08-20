package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// ForgotPasswordProcess : details about the password to update.
type ForgotPasswordProcess struct {
	ID     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	UserID bson.ObjectId `json:"user_id" bson:"userId"`
	Code   string        `json:"code" bson:"code"`
	Tries  uint          `json:"-" bson:"tries"`
}

// NewForgotPasswordProcess : returns a new 'ForgotPasswordProcess' struct.
func NewForgotPasswordProcess(userID string) *ForgotPasswordProcess {
	return &ForgotPasswordProcess{
		UserID: bson.ObjectIdHex(userID),
		Tries:  0,
	}
}

// IsSaved : returns true if the 'ForgotPasswordProcess' document if saved in the collection.
func (fpp *ForgotPasswordProcess) IsSaved() bool {
	return fpp.ID.Valid()
}

// Save : inserts the 'ForgotPasswordProcess' document in the collection, returns an error if needed.
func (fpp *ForgotPasswordProcess) Save(db *mgo.Database) (string, error) {
	return document.SaveDocument(fpp, ForgotPasswordProcessCollection, db)
}

// Update : updates the 'ForgotPasswordProcess' document, returns an error if needed.
func (fpp *ForgotPasswordProcess) Update(update interface{}, db *mgo.Database) error {
	return document.UpdateDocument(fpp.ID, ForgotPasswordProcessCollection, update, db)
}

// Delete : deletes the 'ForgotPasswordProcess' document from the collection, returns an error if needed.
func (fpp *ForgotPasswordProcess) Delete(db *mgo.Database) error {
	return document.DeleteDocument(fpp.ID, ForgotPasswordProcessCollection, db)
}

// GetByID : initializes the 'ForgotPasswordProcess' struct from the 'user' document with the ID.
func (fpp *ForgotPasswordProcess) GetByID(id string, db *mgo.Database) error {
	return document.GetDocumentByID(fpp, ForgotPasswordProcessCollection, id, db)
}

// GetID : returns the 'ForgotPasswordProcess' ID.
func (fpp *ForgotPasswordProcess) GetID() string {
	return fpp.ID.Hex()
}

// SetID : initializes the 'ForgotPasswordProcess' ID.
func (fpp *ForgotPasswordProcess) SetID(id string) {
	fpp.ID = bson.ObjectIdHex(id)
}
