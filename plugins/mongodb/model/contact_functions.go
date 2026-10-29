package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/document"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetContactsForAdmin : returns all the contacts.
func GetContactsForAdmin(ctx context.Context, offset, limit int64, db *mongo.Database) ([]AdminSummarizedContact, error) {
	contacts := []Contact{}
	err := document.GetDocumentsBySortWithOffsetAndLimit(ctx, &contacts, ContactCollection, []string{"-createdAt"}, offset, limit, db)

	adminContacts := make([]AdminSummarizedContact, len(contacts))
	for i, contact := range contacts {
		adminContact := NewAdminSummarizedContact(&contact)
		adminContacts[i] = adminContact
	}

	return adminContacts, err
}
