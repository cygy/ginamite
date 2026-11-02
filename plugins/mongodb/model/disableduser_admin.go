package model

import (
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
)

// AdminDisabledUser is the view of a DisabledUser exposed to the administrators.
type AdminDisabledUser struct {
	ID           bson.ObjectID    `json:"id" bson:"_id,omitempty"`
	Username     string           `json:"username" bson:"username"`
	Slug         string           `json:"slug" bson:"slug"`
	PrivateInfos UserPrivateInfos `json:"private_infos" bson:"privateInfos,omitempty"`
	Verified     bool             `json:"verified" bson:"verified"`
}

// NewAdminDisabledUser returns a new AdminDisabledUser.
func NewAdminDisabledUser(user *User) *AdminDisabledUser {
	adminUser := &AdminDisabledUser{
		Username:     user.Username,
		Slug:         user.Slug,
		PrivateInfos: user.PrivateInfos,
		Verified:     user.Verified,
	}

	return adminUser
}

// URLParams returns the parameters to build the URL of the document: its ID.
func (user AdminDisabledUser) URLParams() map[string]string {
	return map[string]string{
		"id": user.ID.Hex(),
	}
}
