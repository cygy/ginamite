package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// GetEnabledUsersForAdmin : returns the enabled users for the admin.
func GetEnabledUsersForAdmin(ctx context.Context, db store.Database) ([]AdminUser, error) {
	users := []AdminUser{}
	err := document.GetDocumentsBySort(ctx, &users, UserCollection, []string{"username"}, db)

	return users, err
}
