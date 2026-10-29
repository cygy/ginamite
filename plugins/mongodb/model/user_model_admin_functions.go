package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/document"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetEnabledUsersForAdmin : returns the enabled users for the admin.
func GetEnabledUsersForAdmin(ctx context.Context, db *mongo.Database) ([]AdminUser, error) {
	users := []AdminUser{}
	err := document.GetDocumentsBySort(ctx, &users, UserCollection, []string{"username"}, db)

	return users, err
}
