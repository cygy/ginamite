package model

import "github.com/cygy/ginamite/api/rights"

// Settings are the settings of the plugin, used by the model and by the functions of the API.
type Settings struct {
	// UserMaxValidateRegistrationAttemps is the maximum number of attempts to validate a registration.
	UserMaxValidateRegistrationAttemps uint

	// UserMaxDescriptionLength is the maximum length of the description of an account.
	UserMaxDescriptionLength uint

	// UserMaxImageSize is the maximum size of the image of an account, in bytes.
	UserMaxImageSize int64

	// UserImagePath is the path of the directory containing the user images.
	UserImagePath string

	// UserImageWidth is the width of the user images, in pixels.
	UserImageWidth int64

	// UserImageHeight is the height of the user images, in pixels.
	UserImageHeight int64

	// UserImageSmallVersion reports whether the small version of the user images must be generated.
	UserImageSmallVersion bool

	// UserImageThumbVersion reports whether the thumbnail version of the user images must be generated.
	UserImageThumbVersion bool

	// AllUserAbilities lists all the abilities a user can have.
	AllUserAbilities []string

	// DefaultUserAbilities lists the abilities given to a new user.
	DefaultUserAbilities []string
}

// DefaultSettings returns the default settings of the plugin.
func DefaultSettings() Settings {
	return Settings{
		UserMaxValidateRegistrationAttemps: 5,
		UserMaxDescriptionLength:           300,
		UserMaxImageSize:                   2000000,
		UserImagePath:                      "/images",
		UserImageWidth:                     800,
		UserImageHeight:                    800,
		UserImageSmallVersion:              true,
		UserImageThumbVersion:              true,
		AllUserAbilities:                   []string{rights.AbilityToManageUsers, rights.AbilityToManageContacts},
		DefaultUserAbilities:               []string{},
	}
}
