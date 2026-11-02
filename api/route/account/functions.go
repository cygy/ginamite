package account

import (
	"github.com/cygy/ginamite/api/validator"

	"github.com/gin-gonic/gin"
)

// functionsKey is the key of the functions in the Gin context.
const functionsKey = "_account_functions"

// Functions are the functions provided by the application, such as the ones of the storage plugin, used by the
// handlers of the package. A missing function disables the feature relying on it.
type Functions struct {
	// NewTokenValidator returns the validator of the tokens issued by a third-party authentication method. The default
	// one validates the tokens of Facebook and Google with their API.
	NewTokenValidator func(c *gin.Context, method string) validator.TokenValidator

	// Notifications is the storage of the notifications of the users.
	Notifications NotificationsStorage

	// Tokens is the storage of the auth tokens.
	Tokens TokensStorage

	// Processes is the storage of the processes bound to a user.
	Processes ProcessesStorage

	// GetUserDetailsByIdentifier is a function to get the user details from its identifier (username, email addresse, what you want.)
	GetUserDetailsByIdentifier GetUserDetailsByIdentifierFunc

	// GetUserDetailsByID is a function to get the user details from its unique identifier.
	GetUserDetailsByID GetUserDetailsByIDFunc

	// GetUserDetailsByFacebookUserID is a function to get the user details from its unique facebook identifier.
	GetUserDetailsByFacebookUserID GetUserDetailsBySocialNetworkIDFunc

	// GetUserDetailsByGoogleUserID is a function to get the user details from its unique facebook identifier.
	GetUserDetailsByGoogleUserID GetUserDetailsBySocialNetworkIDFunc

	// GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser is a function to get the user details from its identifier (username, email addresse, what you want.)
	GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUserFunc

	// UpdateEncryptedPasswordByUserID is an optional function to replace the encrypted password of a user. If it is
	// defined, the password of a user signing in is encrypted again when its hash is weaker than the current one.
	UpdateEncryptedPasswordByUserID UpdateEncryptedPasswordByUserIDFunc

	// UpdateUserPassword is a function to update the password of an user.
	UpdateUserPassword UpdateUserPasswordFunc

	// UpdateUserEmailAddress is a function to update the email address of an user.
	UpdateUserEmailAddress UpdateUserEmailAddressFunc

	// UserEmailAddressIsVerified is a function to flag the email address of a user as valid.
	UserEmailAddressIsVerified UserEmailAddressIsVerifiedFunc

	// UpdateUserSocialNetworks updates the social network accounts of the user.
	UpdateUserSocialNetworks UpdateUserSocialNetworksFunc

	// ValidateUserRegistration is a function to validate the registration of a user by its ID.
	ValidateUserRegistration ValidateUserRegistrationFunc

	// CancelUserRegistration is a function to cancel the registration of a user by its ID.
	CancelUserRegistration CancelUserRegistrationFunc

	// GetRegistrationDetails is a function to get the details of the registration of a user.
	GetRegistrationDetails GetRegistrationDetailsFunc

	// DoesAccountWithUsernameExist returns true if an account with this username does exist already.
	DoesAccountWithUsernameExist DoesAccountWithUsernameExistFunc

	// DoesAccountWithEmailAddressExist returns true if an account with this email address does exist already.
	DoesAccountWithEmailAddressExist DoesAccountWithEmailAddressExistFunc

	// RegisterByEmailAddress saves a newly registered user.
	RegisterByEmailAddress RegisterByEmailAddressFunc

	// RegisterByThirdPartyToken saves a newly registered user.
	RegisterByThirdPartyToken RegisterByThirdPartyTokenFunc

	// DoesUserWithFacebookUserIDExist returns true if a user with a facebook user id is registered already.
	DoesUserWithFacebookUserIDExist DoesUserWithSocialNetworkUserIDExistFunc

	// DoesUserWithGoogleUserIDExist returns true if a user with a google user id is registered already.
	DoesUserWithGoogleUserIDExist DoesUserWithSocialNetworkUserIDExistFunc

	// SaveFacebookUserTokenDetailsToUser is a function to save the auth token details from facebook to an account.
	SaveFacebookUserTokenDetailsToUser SaveSocialNetworkUserTokenDetailsToUserFunc

	// SaveGoogleUserTokenDetailsToUser is a function to save the auth token details from google to an account.
	SaveGoogleUserTokenDetailsToUser SaveSocialNetworkUserTokenDetailsToUserFunc

	// GetOwnedAccountDetails is a function to get the user details from its unique identifier.
	GetOwnedAccountDetails GetOwnedAccountDetailsFunc

	// UpdatedAccountInfos is a function to update the user infos.
	UpdatedAccountInfos UpdatedAccountInfosFunc

	// UpdatedAccountSettings is a function to update the user settings.
	UpdatedAccountSettings UpdatedAccountSettingsFunc

	// DeleteUserByID deletes a user by its ID.
	DeleteUserByID DeleteUserByIDFunc

	// DisableUserByID disables a user by its ID.
	DisableUserByID DisableUserByIDFunc

	// EnableUserByID enables a user by its ID.
	EnableUserByID EnableUserByIDFunc
}

// Inject returns a middleware saving the functions in the Gin context, see FunctionsOf.
func Inject(functions *Functions) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetFunctions(c, functions)
		c.Next()
	}
}

// SetFunctions saves the functions in the Gin context, like Inject does. It is meant for the tests.
func SetFunctions(c *gin.Context, functions *Functions) {
	c.Set(functionsKey, functions)
}

// FunctionsOf returns the functions saved in the Gin context by Inject.
// It panics if there are none: the handlers can not work without them.
func FunctionsOf(c *gin.Context) *Functions {
	value, ok := c.Get(functionsKey)
	if !ok {
		panic("account: no functions in the context, add the middleware Inject")
	}

	return value.(*Functions)
}
