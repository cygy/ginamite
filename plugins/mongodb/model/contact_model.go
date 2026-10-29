package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Contact : contact properties.
type Contact struct {
	ID           bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
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
	return bson.IsObjectIDProvided(contact.ID)
}

// Save : inserts the 'Contact' document in the collection, returns an error if needed.
func (contact *Contact) Save(ctx context.Context, db *mongo.Database) (string, error) {
	return document.SaveDocument(ctx, contact, ContactCollection, db)
}

// Update : updates the 'Contact' document, returns an error if needed.
func (contact *Contact) Update(ctx context.Context, update any, db *mongo.Database) error {
	return document.UpdateDocument(ctx, contact.ID, ContactCollection, update, db)
}

// Delete : deletes the 'Contact' document from the collection, returns an error if needed.
func (contact *Contact) Delete(ctx context.Context, db *mongo.Database) error {
	return document.DeleteDocument(ctx, contact.ID, ContactCollection, db)
}

// GetByID : initializes the 'Contact' struct from the 'Contact' document with the ID.
func (contact *Contact) GetByID(ctx context.Context, id string, db *mongo.Database) error {
	return document.GetDocumentByID(ctx, contact, ContactCollection, id, db)
}

// GetID : returns the 'Contact' ID.
func (contact *Contact) GetID() string {
	return contact.ID.Hex()
}

// SetID : initializes the 'Contact' ID.
func (contact *Contact) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	contact.ID = objectID
	return nil
}
