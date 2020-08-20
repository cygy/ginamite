package authentication

const (
	// TypeWeb : type of a device
	TypeWeb = "web"

	// TypeIOS : type of a device
	TypeIOS = "ios"

	// TypeAndroid : type of a device
	TypeAndroid = "android"

	// MethodPassword : authentication method by password
	MethodPassword = "password"

	// MethodFacebook : authentication method by facebook
	MethodFacebook = "facebook"

	// MethodGoogle : authentication method by google
	MethodGoogle = "google"

	// InvalidCharactersSetForUsername : forbidden characters in an username.
	InvalidCharactersSetForUsername = "'‘{}"

	// MinimumCharactersForPassword : minimum number of characters in a password.
	MinimumCharactersForPassword = 10

	// MinimumCharactersForUsername : minimum number of characters in a username.
	MinimumCharactersForUsername = 4

	// PrivateKeyLength : length of the generated key code.
	PrivateKeyLength = 128

	// RegisterCodeLength : length of the generated register code.
	RegisterCodeLength = 64

	// PasswordEncryptionSalt : salt used to encrypt the password.
	PasswordEncryptionSalt = "$6$rounds=5000$"

	// PasswordEncryptionSaltLength : length of the salt used to encrypt the password.
	PasswordEncryptionSaltLength = 16

	// ProcessUpdateEmailAddress :
	ProcessUpdateEmailAddress = "update_email"

	// ProcessUpdatePassword :
	ProcessUpdatePassword = "update_password"

	// ProcessVerifyEmailAddress :
	ProcessVerifyEmailAddress = "verify_email"

	// ProcessDeleteAccount :
	ProcessDeleteAccount = "delete_account"

	// ProcessDisableAccount :
	ProcessDisableAccount = "disable_account"
)
