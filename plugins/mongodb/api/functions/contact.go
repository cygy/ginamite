package functions

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// Contacts stores the contacts in a database. It is a contact.Storage. Create it with NewContacts.
type Contacts struct {
	database func() store.Database
}

// NewContacts returns the storage of the contacts working on the database returned by the function, which is called
// each time the storage needs the database.
func NewContacts(database func() store.Database) *Contacts {
	return &Contacts{database: database}
}

// GetContacts returns the contacts for the admin panel.
func (s *Contacts) GetContacts(ctx context.Context, offset, limit int64) (any, error) {
	return model.GetContactsForAdmin(ctx, offset, limit, s.database())
}

// GetContact gets a contact.
func (s *Contacts) GetContact(ctx context.Context, contactID string) (any, error) {
	contact := &model.Contact{}
	if err := contact.GetByID(ctx, contactID, s.database()); err != nil {
		return nil, err
	}

	return model.NewAdminContact(contact), nil
}

// UpdateContact updates a contact.
func (s *Contacts) UpdateContact(ctx context.Context, contactID string, done bool) error {
	db := s.database()

	contact := &model.Contact{}
	if err := contact.GetByID(ctx, contactID, db); err != nil {
		return err
	}

	return contact.UpdateDone(ctx, done, db)
}

// CreateContact creates a contact.
func (s *Contacts) CreateContact(ctx context.Context, emailAddress, subject, text string) error {
	contact := model.NewContact()
	contact.EmailAddress = emailAddress
	contact.Subject = subject
	contact.Text = text

	_, err := contact.Save(ctx, s.database())

	return err
}
