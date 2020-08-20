package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
)

// GetEnabledUsersForAdmin : returns the enabled users for the admin.
func GetEnabledUsersForAdmin(db *mgo.Database) ([]AdminUser, error) {
	users := []AdminUser{}
	err := document.GetDocumentsBySort(&users, UserCollection, []string{"username"}, db)

	return users, err
}
