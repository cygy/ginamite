package model

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// SitemapUser is the view of a User used to build the XML sitemap files.
type SitemapUser struct {
	ID   bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Slug string        `json:"slug" bson:"slug"`
}

// NewSitemapUser returns a new SitemapUser.
func NewSitemapUser(user User) SitemapUser {
	return SitemapUser{
		ID:   user.ID,
		Slug: user.Slug,
	}
}

// GetSitemapUsers returns users.
func GetSitemapUsers(ctx context.Context, offset, limit int64, db store.Database) ([]SitemapUser, error) {
	users := []User{}
	err := document.GetDocumentsBySortWithOffsetAndLimit(ctx, &users, UserCollection, []string{"registrationInfos.date"}, offset, limit, db)

	sitemapUsers := make([]SitemapUser, len(users))
	for i, user := range users {
		sitemapUsers[i] = NewSitemapUser(user)
	}

	return sitemapUsers, err
}
