package model

import (
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/globalsign/mgo"
)

// GetSitemapUsers : returns users.
func GetSitemapUsers(offset, limit int, db *mgo.Database) ([]SitemapUser, error) {
	users := []User{}
	err := document.GetDocumentsBySortWithOffsetAndLimit(&users, UserCollection, []string{"registrationInfos.date"}, offset, limit, db)

	sitemapUsers := make([]SitemapUser, len(users))
	for i, user := range users {
		sitemapUsers[i] = NewSitemapUser(user)
	}

	return sitemapUsers, err
}
