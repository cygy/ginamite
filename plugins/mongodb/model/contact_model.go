package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Contact : contact properties.
type Contact struct {
	ID           bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	EmailAddress string        `json:"email_address" bson:"emailAddress"`
	Subject      string        `json:"subject" bson:"subject"`
	Text         string        `json:"text" bson:"text"`
	Done         bool          `json:"done" bson:"done"`
	CreatedAt    time.Time     `json:"created_at" bson:"createdAt"`
}

// NewContact : returns a new 'Contact' struct.
func NewContact() *Contact {
	return &Contact{
		Done:      false,
		CreatedAt: time.Now(),
	}
}

// IsSaved : returns true if the 'Contact' document if saved in the collection.
func (contact *Contact) IsSaved() bool {
	return contact.ID.Valid()
}

// Save : inserts the 'Contact' document in the collection, returns an error if needed.
func (contact *Contact) Save(db *mgo.Database) (string, error) {
	return document.SaveDocument(contact, ContactCollection, db)
}

// Update : updates the 'Contact' document, returns an error if needed.
func (contact *Contact) Update(update interface{}, db *mgo.Database) error {
	return document.UpdateDocument(contact.ID, ContactCollection, update, db)
}

// Delete : deletes the 'Contact' document from the collection, returns an error if needed.
func (contact *Contact) Delete(db *mgo.Database) error {
	return document.DeleteDocument(contact.ID, ContactCollection, db)
}

// GetByID : initializes the 'Contact' struct from the 'Contact' document with the ID.
func (contact *Contact) GetByID(id string, db *mgo.Database) error {
	return document.GetDocumentByID(contact, ContactCollection, id, db)
}

// GetID : returns the 'Contact' ID.
func (contact *Contact) GetID() string {
	return contact.ID.Hex()
}

// SetID : initializes the 'Contact' ID.
func (contact *Contact) SetID(id string) {
	contact.ID = bson.ObjectIdHex(id)
}
