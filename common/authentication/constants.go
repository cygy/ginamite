// Package authentication defines the authentication primitives shared by the
// API: devices, users, auth tokens (JWT) and password encryption.
package authentication

// Types of device.
const (
	// TypeWeb is the type of a web browser.
	TypeWeb = "web"

	// TypeIOS is the type of an iOS device.
	TypeIOS = "ios"

	// TypeAndroid is the type of an Android device.
	TypeAndroid = "android"
)

// Authentication methods.
const (
	// MethodPassword is the authentication by password.
	MethodPassword = "password"

	// MethodFacebook is the authentication by Facebook.
	MethodFacebook = "facebook"

	// MethodGoogle is the authentication by Google.
	MethodGoogle = "google"
)

const (
	// InvalidCharactersSetForUsername lists the characters forbidden in a username.
	InvalidCharactersSetForUsername = "'‘{}/"

	// MinimumCharactersForPassword is the minimum number of characters in a password.
	MinimumCharactersForPassword = 10

	// MinimumCharactersForUsername is the minimum number of characters in a username.
	MinimumCharactersForUsername = 4

	// PrivateKeyLength is the length of the generated private key.
	PrivateKeyLength = 128

	// RegisterCodeLength is the length of the generated register code.
	RegisterCodeLength = 64

	// PasswordEncryptionSalt is the salt prefix used to encrypt a password.
	PasswordEncryptionSalt = "$6$rounds=5000$"

	// PasswordEncryptionSaltLength is the length of the random part of the salt used to encrypt a password.
	PasswordEncryptionSaltLength = 16
)

// Types of process a user can start on their account.
const (
	// ProcessUpdateEmailAddress is the process to update the email address.
	ProcessUpdateEmailAddress = "update_email"

	// ProcessUpdatePassword is the process to update the password.
	ProcessUpdatePassword = "update_password"

	// ProcessVerifyEmailAddress is the process to verify the email address.
	ProcessVerifyEmailAddress = "verify_email"

	// ProcessDeleteAccount is the process to delete the account.
	ProcessDeleteAccount = "delete_account"

	// ProcessDisableAccount is the process to disable the account.
	ProcessDisableAccount = "disable_account"
)
