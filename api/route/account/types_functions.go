package account

import (
	"time"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// NotificationsReceivedByUserFunc : function to get all the notifications received by the user.
type NotificationsReceivedByUserFunc func(c *gin.Context, userID string) (interface{}, error)

// NotificationReceivedByUserFunc : function to get a notification received by the user.
type NotificationReceivedByUserFunc func(c *gin.Context, userID, notificationID string) (interface{}, error)

// UpdateUserNotificationReadFunc : function to update the read flag of a notification received by the user.
type UpdateUserNotificationReadFunc func(c *gin.Context, userID, notificationID string) error

// UpdateUserNotificationNotifiedFunc : function to update the notified flag of a notification received by the user.
type UpdateUserNotificationNotifiedFunc func(c *gin.Context, userID, notificationID string) error

// DeleteUserNotificationFunc : function to delete a notification received by the user.
type DeleteUserNotificationFunc func(c *gin.Context, userID, notificationID string) error

// NotificationsSubscribedByUserFunc : function to get all the notifications subscribed by the user.
type NotificationsSubscribedByUserFunc func(c *gin.Context, userID string) []string

// SubscribeNotificationsToUserFunc : function to subscribe some notifications to the user.
type SubscribeNotificationsToUserFunc func(c *gin.Context, userID string, notifications []string) error

// UnsubscribeNotificationFromUserFunc : function to unsubscribe a notification from a user.
type UnsubscribeNotificationFromUserFunc func(c *gin.Context, userID, notification, unsubscribeKey string) error

// GetUserDetailsByIdentifierFunc : function to get the user details from its identifier (username, email addresse, what you want.)
type GetUserDetailsByIdentifierFunc func(c *gin.Context, identifier string) (authentication.User, error)

// GetUserDetailsByIDFunc : function to get the user details from its unique identifier.
type GetUserDetailsByIDFunc func(c *gin.Context, userID string) (authentication.User, error)

// GetUserDetailsBySocialNetworkIDFunc : function to get the user details from its unique social network identifier.
type GetUserDetailsBySocialNetworkIDFunc func(c *gin.Context, userID string) (authentication.User, error)

// GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUserFunc : function to get the user details and its encrypted password from its identifier (username, email addresse, what you want.)
type GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUserFunc func(c *gin.Context, identifier string) (authentication.User, string, error)

// NewForgotPasswordProcessForUserFunc : function to create a new forgot password process for a user. Return the unique identifier of the process.
type NewForgotPasswordProcessForUserFunc func(c *gin.Context, userID, processPrivateKey string) (string, error)

// GetForgotPasswordProcessDetailsByIDFunc : function to get the forgot password process details from its identifier.
type GetForgotPasswordProcessDetailsByIDFunc func(c *gin.Context, processID string) (userID, processPrivateKey string, err error)

// WrongAccessToForgotPasswordProcessFunc : function to avoid brute force, increment tries or delete the process.
type WrongAccessToForgotPasswordProcessFunc func(c *gin.Context, processID string)

// UpdateUserPasswordFunc : function to update the password of an user.
type UpdateUserPasswordFunc func(c *gin.Context, userID, password string, isPasswordEncrypted bool) error

// UpdateUserEmailAddressFunc : function to update the email address of an user.
type UpdateUserEmailAddressFunc func(c *gin.Context, userID, emailAddress string) error

// UserEmailAddressIsVerifiedFunc : function to flag the email address of a user as valid.
type UserEmailAddressIsVerifiedFunc func(c *gin.Context, userID string) error

// UpdateUserSocialNetworksFunc : updates the social network accounts of the user.
type UpdateUserSocialNetworksFunc func(c *gin.Context, userID, facebook, twitter, instagram string) error

// DeleteForgotPasswordProcessFunc : function to delete a forgot password process by ID.
type DeleteForgotPasswordProcessFunc func(c *gin.Context, processID string) error

// ValidateUserRegistrationFunc : function to validate the registration of a user by its ID.
type ValidateUserRegistrationFunc func(c *gin.Context, userID, registrationKey string) error

// CancelUserRegistrationFunc : function to cancel the registration of a user by its ID.
type CancelUserRegistrationFunc func(c *gin.Context, userID, registrationKey string) error

// GetRegistrationDetailsFunc : function to get the details of the registration of a user.
type GetRegistrationDetailsFunc func(c *gin.Context, userID string) (*authentication.Details, error)

// SaveAuthenticationTokenFunc : function to save the authentication details and return the unique ID.
type SaveAuthenticationTokenFunc func(c *gin.Context, details authentication.Details) (tokenID string, knownIPAddress bool, err error)

// DoesAccountWithUsernameExistFunc : returns true if an account with this username does exist already.
type DoesAccountWithUsernameExistFunc func(c *gin.Context, username string) bool

// DoesAccountWithEmailAddressExistFunc : returns true if an account with this email address does exist already.
type DoesAccountWithEmailAddressExistFunc func(c *gin.Context, emailAddress string) bool

// RegisterByEmailAddressFunc : saves a newly registered user.
type RegisterByEmailAddressFunc func(c *gin.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error

// RegisterByThirdPartyTokenFunc : saves a newly registered user.
type RegisterByThirdPartyTokenFunc func(c *gin.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error

// DoesUserWithSocialNetworkUserIDExistFunc : function to know if a user with a social network user id is registered already.
type DoesUserWithSocialNetworkUserIDExistFunc func(c *gin.Context, userID string) bool

// SaveSocialNetworkUserTokenDetailsToUserFunc : function to save the auth token details from a social network to an account.
type SaveSocialNetworkUserTokenDetailsToUserFunc func(c *gin.Context, userID string, tokenInfos validator.TokenInfos) error

// GetTokenExpirationDateFunc : function to get the expiration date of a token from its unique identifier.
type GetTokenExpirationDateFunc func(c *gin.Context, tokenID string) (time.Time, error)

// DeleteTokenByIDFunc : delete an auth token by its ID.
type DeleteTokenByIDFunc func(c *gin.Context, tokenID string) error

// GetTokenDetailsByIDFunc : function to get the token details from its unique identifier.
type GetTokenDetailsByIDFunc func(c *gin.Context, tokenID string) (authentication.Token, error)

// GetOwnedTokensFunc : function to get the auth tokens of a user.
type GetOwnedTokensFunc func(c *gin.Context, userID string) (interface{}, error)

// GetOwnedTokenByIDFunc : function to get the auth token by its ID and the ID of the token's owner.
type GetOwnedTokenByIDFunc func(c *gin.Context, tokenID string) (token interface{}, ownerID string, err error)

// UpdateTokenByIDFunc : function to update the properties of a token.
type UpdateTokenByIDFunc func(c *gin.Context, tokenID, name string, enableNotifications bool, notifications []string) error

// GetOwnedAccountDetailsFunc : function to get the user details from its unique identifier.
type GetOwnedAccountDetailsFunc func(c *gin.Context, userID string) interface{}

// UpdatedAccountInfosFunc : function to update the user infos.
type UpdatedAccountInfosFunc func(c *gin.Context, userID, firstName, lastName string) error

// UpdatedAccountSettingsFunc : function to update the user settings.
type UpdatedAccountSettingsFunc func(c *gin.Context, userID, locale, timezone string) error

// CreateNewProcessFunc : function to start a new process bound to a user.
type CreateNewProcessFunc func(c *gin.Context, userID string) (processID, processKey string, err error)

// CreateNewProcessWithValueFunc : function to start a new process bound to a user.
type CreateNewProcessWithValueFunc func(c *gin.Context, userID, value string) (processID, processKey string, err error)

// VerifyProcessFunc : function to verify a process and get the ID of the bound user.
type VerifyProcessFunc func(c *gin.Context, processID, key string) (userID, newValue string, err error)

// DeleteProcessByIDFunc : deletes a process by its ID.
type DeleteProcessByIDFunc func(c *gin.Context, processID string) error

// DeleteUserByIDFunc : deletes a user by its ID.
type DeleteUserByIDFunc func(c *gin.Context, userID string) error

// DisableUserByIDFunc : disables a user by its ID.
type DisableUserByIDFunc func(c *gin.Context, userID string) (authentication.User, error)

// EnableUserByIDFunc : enables a user by its ID.
type EnableUserByIDFunc func(c *gin.Context, userID string) (authentication.User, error)
