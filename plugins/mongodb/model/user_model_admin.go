package model

import "github.com/globalsign/mgo/bson"

// AdminUser : admin version of the User struct
type AdminUser struct {
	ID           bson.ObjectId         `json:"id" bson:"_id,omitempty"`
	Username     string                `json:"username" bson:"username"`
	PrivateInfos AdminUserPrivateInfos `json:"private_infos" bson:"privateInfos,omitempty"`
	Verified     bool                  `json:"verified" bson:"verified"`
}

// AdminUserPrivateInfos : admin version of the UserPrivateInfos struct
type AdminUserPrivateInfos struct {
	Email        string `json:"email" bson:"email"`
	IsEmailValid bool   `json:"is_email_valid" bson:"isEmailValid"`
}

// NewAdminUser : returns a new 'AdminUser' struct.
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
