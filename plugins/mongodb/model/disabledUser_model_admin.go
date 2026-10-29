package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AdminDisabledUser : properties of a disabled user for the admin.
type AdminDisabledUser struct {
	ID           bson.ObjectID    `json:"id" bson:"_id,omitempty"`
	Username     string           `json:"username" bson:"username"`
	Slug         string           `json:"slug" bson:"slug"`
	PrivateInfos UserPrivateInfos `json:"private_infos" bson:"privateInfos,omitempty"`
	Verified     bool             `json:"verified" bson:"verified"`
}

// NewAdminDisabledUser : returns a new 'AdminDisabledUser' struct.
func NewAdminDisabledUser(user *User) *AdminDisabledUser {
	adminUser := &AdminDisabledUser{
		Username:     user.Username,
		Slug:         user.Slug,
		PrivateInfos: user.PrivateInfos,
		Verified:     user.Verified,
	}

	return adminUser
}
