package model

import (
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/globalsign/mgo"
)

// GetContactsForAdmin : returns all the contacts.
func GetContactsForAdmin(offset, limit int, db *mgo.Database) ([]AdminSummarizedContact, error) {
	contacts := []Contact{}
	err := document.GetDocumentsBySortWithOffsetAndLimit(&contacts, ContactCollection, []string{"-createdAt"}, offset, limit, db)

	adminContacts := make([]AdminSummarizedContact, len(contacts))
	for i, contact := range contacts {
		adminContact := NewAdminSummarizedContact(&contact)
		adminContacts[i] = adminContact
	}

	return adminContacts, err
}
