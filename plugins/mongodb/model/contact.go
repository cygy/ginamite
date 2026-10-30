package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// ContactCollection is the name of the collection of the contact documents.
const ContactCollection = "contact"

// Contact is a message sent through the contact form.
type Contact struct {
	ID           bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	EmailAddress string        `json:"email_address" bson:"emailAddress"`
	Subject      string        `json:"subject" bson:"subject"`
	Text         string        `json:"text" bson:"text"`
	Done         bool          `json:"done" bson:"done"`
	CreatedAt    time.Time     `json:"created_at" bson:"createdAt"`
}

// NewContact returns a new Contact.
func NewContact() *Contact {
	return &Contact{
		Done:      false,
		CreatedAt: time.Now(),
	}
}

// IsSaved reports whether the Contact document is saved in the collection.
func (contact *Contact) IsSaved() bool {
	return bson.IsObjectIDProvided(contact.ID)
}

// Save inserts the Contact document in the collection, returning an error if needed.
func (contact *Contact) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, contact, ContactCollection, db)
}

// Update updates the Contact document, returning an error if needed.
func (contact *Contact) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, contact.ID, ContactCollection, update, db)
}

// Delete deletes the Contact document from the collection, returning an error if needed.
func (contact *Contact) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, contact.ID, ContactCollection, db)
}

// GetByID loads the Contact document with the ID.
func (contact *Contact) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, contact, ContactCollection, id, db)
}

// GetID returns the Contact ID.
func (contact *Contact) GetID() string {
	return contact.ID.Hex()
}

// SetID sets the ID of the Contact from its hexadecimal form. The ID is kept if the form is not valid.
func (contact *Contact) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	contact.ID = objectID
	return nil
}

// UpdateDone updates the done property of a contact.
func (contact *Contact) UpdateDone(ctx context.Context, done bool, db store.Database) error {
	return contact.Update(ctx, bson.M{"$set": bson.M{"done": done}}, db)
}

// GetContactsForAdmin returns all the contacts.
func GetContactsForAdmin(ctx context.Context, offset, limit int64, db store.Database) ([]AdminSummarizedContact, error) {
	contacts := []Contact{}
	err := document.GetDocumentsBySortWithOffsetAndLimit(ctx, &contacts, ContactCollection, []string{"-createdAt"}, offset, limit, db)

	adminContacts := make([]AdminSummarizedContact, len(contacts))
	for i, contact := range contacts {
		adminContact := NewAdminSummarizedContact(&contact)
		adminContacts[i] = adminContact
	}

	return adminContacts, err
}
