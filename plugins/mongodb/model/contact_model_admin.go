package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// AdminContact : admin version of the Contact struct
type AdminContact Contact

// NewAdminContact : returns a new 'AdminContact' struct.
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

// AdminSummarizedContact : short version of a contact document.
type AdminSummarizedContact struct {
	ID           bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	EmailAddress string        `json:"email_address" bson:"emailAddress"`
	Subject      string        `json:"subject" bson:"subject"`
	Done         bool          `json:"done" bson:"done"`
	CreatedAt    time.Time     `json:"created_at" bson:"createdAt"`
}

// NewAdminSummarizedContact : returns a new 'AdminSummarizedContact' struct.
func NewAdminSummarizedContact(contact *Contact) AdminSummarizedContact {
	return AdminSummarizedContact{
		ID:           contact.ID,
		EmailAddress: contact.EmailAddress,
		Subject:      contact.Subject,
		Done:         contact.Done,
		CreatedAt:    contact.CreatedAt,
	}
}
