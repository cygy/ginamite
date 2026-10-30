package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
)

// AdminContact is the view of a Contact exposed to the administrators.
type AdminContact Contact

// NewAdminContact returns a new AdminContact.
func NewAdminContact(contact *Contact) *AdminContact {
	return &AdminContact{
		ID:           contact.ID,
		EmailAddress: contact.EmailAddress,
		Subject:      contact.Subject,
		Text:         contact.Text,
		Done:         contact.Done,
		CreatedAt:    contact.CreatedAt,
	}
}

// AdminSummarizedContact is the short view of a Contact exposed to the administrators.
type AdminSummarizedContact struct {
	ID           bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	EmailAddress string        `json:"email_address" bson:"emailAddress"`
	Subject      string        `json:"subject" bson:"subject"`
	Done         bool          `json:"done" bson:"done"`
	CreatedAt    time.Time     `json:"created_at" bson:"createdAt"`
}

// NewAdminSummarizedContact returns a new AdminSummarizedContact.
func NewAdminSummarizedContact(contact *Contact) AdminSummarizedContact {
	return AdminSummarizedContact{
		ID:           contact.ID,
		EmailAddress: contact.EmailAddress,
		Subject:      contact.Subject,
		Done:         contact.Done,
		CreatedAt:    contact.CreatedAt,
	}
}

// URLParams returns the parameters to build the URL of the document: its ID.
func (contact AdminContact) URLParams() map[string]string {
	return map[string]string{
		"id": contact.ID.Hex(),
	}
}

// URLParams returns the parameters to build the URL of the document: its ID.
func (contact AdminSummarizedContact) URLParams() map[string]string {
	return map[string]string{
		"id": contact.ID.Hex(),
	}
}
