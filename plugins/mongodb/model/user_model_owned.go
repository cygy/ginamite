package model

import "github.com/globalsign/mgo/bson"

// OwnUser : public version of the User struct
type OwnUser struct {
	ID                     bson.ObjectId       `json:"id"`
	Username               string              `json:"username"`
	PublicInfos            OwnUserPublicInfos  `json:"public_infos"`
	PrivateInfos           OwnUserPrivateInfos `json:"private_infos"`
	Settings               OwnUserSettings     `json:"settings"`
	Abilities              []string            `json:"abilities"`
	Verified               bool                `json:"verified"`
	VersionOfTermsAccepted string              `json:"version_of_terms_accepted"`
}

// OwnUserPublicInfos : public informations of a user.
type OwnUserPublicInfos struct {
	Description map[string]string `json:"description,omitempty"`
	Facebook    string            `json:"facebook,omitempty"`
	Twitter     string            `json:"twitter,omitempty"`
	Instagram   string            `json:"instagram,omitempty"`
	Image       OwnImage          `json:"image,omitempty"`
}

// OwnUserPrivateInfos : public version of the UserPrivateInfos struct
type OwnUserPrivateInfos struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	IsEmailValid bool   `json:"is_email_valid"`
}

// OwnUserSettings : settings of a user.
type OwnUserSettings struct {
	Locale   string `json:"locale" bson:"locale"`
	Timezone string `json:"timezone" bson:"timezone"`
}

// NewOwnUser : returns a new 'OwnUser' struct.
func NewOwnUser(user *User) *OwnUser {
	ownedUser := &OwnUser{
		ID:       user.ID,
		Username: user.Username,
		PublicInfos: OwnUserPublicInfos{
			Description: user.PublicInfos.Description,
			Facebook:    user.PublicInfos.Facebook,
			Twitter:     user.PublicInfos.Twitter,
			Instagram:   user.PublicInfos.Instagram,
			Image:       NewOwnImage(user.PublicInfos.Image),
		},
		PrivateInfos: OwnUserPrivateInfos{
			FirstName:    user.PrivateInfos.FirstName,
			LastName:     user.PrivateInfos.LastName,
			Email:        user.PrivateInfos.Email,
			IsEmailValid: user.PrivateInfos.IsEmailValid,
		},
		Settings: OwnUserSettings{
			Locale:   user.Settings.Locale,
			Timezone: user.Settings.Timezone,
		},
		Abilities:              user.Abilities,
		Verified:               user.Verified,
		VersionOfTermsAccepted: user.VersionOfTermsAccepted,
	}

	return ownedUser
}

// LocalizedDescription : returns the localized description of a user.
func (publicInfos *OwnUserPublicInfos) LocalizedDescription(locale string) string {
	if localizedDescription, ok := publicInfos.Description[locale]; ok {
		return localizedDescription
	}

	return ""
}
