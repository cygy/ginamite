package model

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// AdminUser is the view of a User exposed to the administrators.
type AdminUser struct {
	ID           bson.ObjectID         `json:"id" bson:"_id,omitempty"`
	Username     string                `json:"username" bson:"username"`
	PrivateInfos AdminUserPrivateInfos `json:"private_infos" bson:"privateInfos,omitempty"`
	Verified     bool                  `json:"verified" bson:"verified"`
}

// AdminUserPrivateInfos is the view of the UserPrivateInfos exposed to the administrators.
type AdminUserPrivateInfos struct {
	Email        string `json:"email" bson:"email"`
	IsEmailValid bool   `json:"is_email_valid" bson:"isEmailValid"`
}

// NewAdminUser returns a new AdminUser.
func NewAdminUser(user *User) *AdminUser {
	adminUser := &AdminUser{
		ID:       user.ID,
		Username: user.Username,
		PrivateInfos: AdminUserPrivateInfos{
			Email:        user.PrivateInfos.Email,
			IsEmailValid: user.PrivateInfos.IsEmailValid,
		},
		Verified: user.Verified,
	}

	return adminUser
}

// GetEnabledUsersForAdmin returns the enabled users for the admin.
func GetEnabledUsersForAdmin(ctx context.Context, db store.Database) ([]AdminUser, error) {
	users := []AdminUser{}
	err := document.GetDocumentsBySort(ctx, &users, UserCollection, []string{"username"}, db)

	return users, err
}
