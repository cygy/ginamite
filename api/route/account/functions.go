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

	// NotificationsReceivedByUser is a function to get all the notifications received by the user.
	NotificationsReceivedByUser NotificationsReceivedByUserFunc

	// NotificationReceivedByUser is a function to get a notification received by the user.
	NotificationReceivedByUser NotificationReceivedByUserFunc

	// UpdateUserNotificationRead is a function to update the read flag of a notification received by the user.
	UpdateUserNotificationRead UpdateUserNotificationReadFunc

	// UpdateUserNotificationNotified is a function to update the notified flag of a notification received by the user.
	UpdateUserNotificationNotified UpdateUserNotificationNotifiedFunc

	// DeleteUserNotification is a function to delete a notification received by the user.
	DeleteUserNotification DeleteUserNotificationFunc

	// NotificationsSubscribedByUser is a function to get all the notifications subscribed by the user.
	NotificationsSubscribedByUser NotificationsSubscribedByUserFunc

	// SubscribeNotificationsToUser is a function to subscribe some notifications to the user.
	SubscribeNotificationsToUser SubscribeNotificationsToUserFunc

	// UnsubscribeNotificationFromUser is a function to unsubscribe a notification from a user.
	UnsubscribeNotificationFromUser UnsubscribeNotificationFromUserFunc

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

	// NewForgotPasswordProcessForUser is a function to create a new forgot password process for a user. Return the unique identifier of the process.
	NewForgotPasswordProcessForUser NewForgotPasswordProcessForUserFunc

	// GetForgotPasswordProcessDetailsByID is a function to get the forgot password process details from its identifier.
	GetForgotPasswordProcessDetailsByID GetForgotPasswordProcessDetailsByIDFunc

	// WrongAccessToForgotPasswordProcess is a function to avoid brute force, increment tries or delete the process.
	WrongAccessToForgotPasswordProcess WrongAccessToForgotPasswordProcessFunc

	// UpdateUserPassword is a function to update the password of an user.
	UpdateUserPassword UpdateUserPasswordFunc

	// UpdateUserEmailAddress is a function to update the email address of an user.
	UpdateUserEmailAddress UpdateUserEmailAddressFunc

	// UserEmailAddressIsVerified is a function to flag the email address of a user as valid.
	UserEmailAddressIsVerified UserEmailAddressIsVerifiedFunc

	// UpdateUserSocialNetworks updates the social network accounts of the user.
	UpdateUserSocialNetworks UpdateUserSocialNetworksFunc

	// DeleteForgotPasswordProcess is a function to delete a forgot password process by ID.
	DeleteForgotPasswordProcess DeleteForgotPasswordProcessFunc

	// ValidateUserRegistration is a function to validate the registration of a user by its ID.
	ValidateUserRegistration ValidateUserRegistrationFunc

	// CancelUserRegistration is a function to cancel the registration of a user by its ID.
	CancelUserRegistration CancelUserRegistrationFunc

	// GetRegistrationDetails is a function to get the details of the registration of a user.
	GetRegistrationDetails GetRegistrationDetailsFunc

	// SaveAuthenticationToken is a function to save the authentication details and return the unique ID.
	SaveAuthenticationToken SaveAuthenticationTokenFunc

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

	// GetTokenExpirationDate is a function to get the expiration date of a token from its unique identifier.
	GetTokenExpirationDate GetTokenExpirationDateFunc

	// DeleteTokenByID delete an auth token by its ID.
	DeleteTokenByID DeleteTokenByIDFunc

	// GetTokenDetailsByID is a function to get the token details from its unique identifier.
	GetTokenDetailsByID GetTokenDetailsByIDFunc

	// GetOwnedTokens is a function to get the auth tokens of a user.
	GetOwnedTokens GetOwnedTokensFunc

	// GetOwnedTokenByID is a function to get the auth token by its ID and the ID of the token's owner.
	GetOwnedTokenByID GetOwnedTokenByIDFunc

	// UpdateTokenByID is a function to update the properties of a token.
	UpdateTokenByID UpdateTokenByIDFunc

	// GetOwnedAccountDetails is a function to get the user details from its unique identifier.
	GetOwnedAccountDetails GetOwnedAccountDetailsFunc

	// UpdatedAccountInfos is a function to update the user infos.
	UpdatedAccountInfos UpdatedAccountInfosFunc

	// UpdatedAccountSettings is a function to update the user settings.
	UpdatedAccountSettings UpdatedAccountSettingsFunc

	// CreateNewDeleteAccountProcess is a function to start a new process to delete an account.
	CreateNewDeleteAccountProcess CreateNewProcessFunc

	// CreateNewDisableAccountProcess is a function to start a new process to disable an account.
	CreateNewDisableAccountProcess CreateNewProcessFunc

	// CreateNewUpdatePasswordProcess is a function to start a new process to update the password of an account.
	CreateNewUpdatePasswordProcess CreateNewProcessWithValueFunc

	// CreateNewUpdateEmailAddressProcess is a function to start a new process to update the email address of an account.
	CreateNewUpdateEmailAddressProcess CreateNewProcessWithValueFunc

	// CreateNewVerifyEmailAddressProcess is a function to start a new process to verify the email address of an account.
	CreateNewVerifyEmailAddressProcess CreateNewProcessFunc

	// VerifyProcess is a function to verify a process and get the ID of the bound user.
	VerifyProcess VerifyProcessFunc

	// DeleteProcessByID deletes a process by its ID.
	DeleteProcessByID DeleteProcessByIDFunc

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
