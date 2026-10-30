package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// GetSitemapUsers : returns users.
func GetSitemapUsers(ctx context.Context, offset, limit int64, db store.Database) ([]SitemapUser, error) {
	users := []User{}
	err := document.GetDocumentsBySortWithOffsetAndLimit(ctx, &users, UserCollection, []string{"registrationInfos.date"}, offset, limit, db)

	sitemapUsers := make([]SitemapUser, len(users))
	for i, user := range users {
		sitemapUsers[i] = NewSitemapUser(user)
	}

	return sitemapUsers, err
}
