package model

// UserCollection : collection name of the 'user' documents.
const UserCollection = "users"

var (
	// UserMaxValidateRegistrationAttemps : maximum attempts to validate a registration.
	UserMaxValidateRegistrationAttemps uint = 5

	// UserMaxDescriptionLength : maximum length of the description of an account.
	UserMaxDescriptionLength uint = 300

	// UserMaxImageSize : maximum size of the image of the account (in bytes).
	UserMaxImageSize int64 = 2000000

	// UserImagePath : path of the directory containing the user images.
	UserImagePath = "/images"

	// UserImageWidth : width of the user images (in px).
	UserImageWidth int64 = 800

	// UserImageHeight : height of the user images (in px).
	UserImageHeight int64 = 800

	// UserImageSmallVersion : true if the small version of the user's image must be generated.
	UserImageSmallVersion = true

	// UserImageThumbVersion : true if the small version of the user's image must be generated.
	UserImageThumbVersion = true
)
