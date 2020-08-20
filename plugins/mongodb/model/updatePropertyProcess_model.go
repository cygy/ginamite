package model

import (
	"fmt"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/random"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UpdatePropertyProcess : details about a property to update.
type UpdatePropertyProcess struct {
	ID     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	UserID bson.ObjectId `json:"user_id" bson:"userId"`
	Type   string        `json:"type" bson:"type"`
	Value  string        `json:"value" bson:"value"`
	Code   string        `json:"code" bson:"code"`
	Tries  uint          `json:"-" bson:"tries"`
	TTLKey string        `json:"-" bson:"userIdtype"` // This field is indexed with a TTL. A TTL index must use only one field, so the pair 'userId / type' can not be used.
}

// NewUpdatePropertyProcess : returns a new 'UpdatePropertyProcess' struct.
func NewUpdatePropertyProcess(userID, t, value string) *UpdatePropertyProcess {
	return &UpdatePropertyProcess{
		UserID: bson.ObjectIdHex(userID),
		Type:   t,
		Value:  value,
		Code:   random.String(64),
		Tries:  0,
		TTLKey: fmt.Sprintf("%s%s", userID, t),
	}
}

// NewUpdatePasswordProcess : returns a new 'UpdatePropertyProcess' struct.
func NewUpdatePasswordProcess(userID, value string) *UpdatePropertyProcess {
	return NewUpdatePropertyProcess(userID, authentication.ProcessUpdatePassword, value)
}

// NewUpdateEmailAddressProcess : returns a new 'UpdatePropertyProcess' struct.
func NewUpdateEmailAddressProcess(userID, value string) *UpdatePropertyProcess {
	return NewUpdatePropertyProcess(userID, authentication.ProcessUpdateEmailAddress, value)
}

// NewUpdateValidEmailAddressProcess : returns a new 'UpdatePropertyProcess' struct.
func NewUpdateValidEmailAddressProcess(userID string) *UpdatePropertyProcess {
	return NewUpdatePropertyProcess(userID, authentication.ProcessVerifyEmailAddress, "true")
}

// NewDeleteAccountProcess : returns a new 'UpdatePropertyProcess' struct.
func NewDeleteAccountProcess(userID string) *UpdatePropertyProcess {
	return NewUpdatePropertyProcess(userID, authentication.ProcessDeleteAccount, "true")
}

// NewDisableAccountProcess : returns a new 'UpdatePropertyProcess' struct.
func NewDisableAccountProcess(userID string) *UpdatePropertyProcess {
	return NewUpdatePropertyProcess(userID, authentication.ProcessDisableAccount, "true")
}

// IsSaved : returns true if the 'UpdatePropertyProcess' document if saved in the collection.
func (process *UpdatePropertyProcess) IsSaved() bool {
	return process.ID.Valid()
}

// Save : inserts the 'UpdatePropertyProcess' document in the collection, returns an error if needed.
func (process *UpdatePropertyProcess) Save(db *mgo.Database) (string, error) {
	return document.SaveDocument(process, UpdatePropertyProcessCollection, db)
}

// Update : updates the 'UpdatePropertyProcess' document, returns an error if needed.
func (process *UpdatePropertyProcess) Update(update interface{}, db *mgo.Database) error {
	return document.UpdateDocument(process.ID, UpdatePropertyProcessCollection, update, db)
}

// Delete : deletes the 'UpdatePropertyProcess' document from the collection, returns an error if needed.
func (process *UpdatePropertyProcess) Delete(db *mgo.Database) error {
	return document.DeleteDocument(process.ID, UpdatePropertyProcessCollection, db)
}

// GetByID : initializes the 'UpdatePropertyProcess' struct from the 'user' document with the ID.
func (process *UpdatePropertyProcess) GetByID(id string, db *mgo.Database) error {
	return document.GetDocumentByID(process, UpdatePropertyProcessCollection, id, db)
}

// GetID : returns the 'UpdatePropertyProcess' ID.
func (process *UpdatePropertyProcess) GetID() string {
	return process.ID.Hex()
}

// SetID : initializes the 'UpdatePropertyProcess' ID.
func (process *UpdatePropertyProcess) SetID(id string) {
	process.ID = bson.ObjectIdHex(id)
}
