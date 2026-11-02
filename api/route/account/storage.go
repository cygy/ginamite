package account

import (
	"context"
	"time"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
)

// NotificationsStorage is the storage of the notifications of the users, provided by the application, such as the one
// of the storage plugin. The context is the one of the request.
type NotificationsStorage interface {
	// NotificationsReceivedByUser returns the latest notifications received by the user, with the texts in the locale.
	NotificationsReceivedByUser(ctx context.Context, userID, locale string) (any, error)
	// NotificationReceivedByUser returns a notification received by the user, with the texts in the locale.
	NotificationReceivedByUser(ctx context.Context, userID, notificationID, locale string) (any, error)
	// UpdateUserNotificationRead marks a notification received by the user as read.
	UpdateUserNotificationRead(ctx context.Context, userID, notificationID string) error
	// UpdateUserNotificationNotified marks a notification received by the user as notified.
	UpdateUserNotificationNotified(ctx context.Context, userID, notificationID string) error
	// DeleteUserNotification deletes a notification received by the user.
	DeleteUserNotification(ctx context.Context, userID, notificationID string) error
	// NotificationsSubscribedByUser returns the notifications subscribed by the user.
	NotificationsSubscribedByUser(ctx context.Context, userID string) ([]string, error)
	// SubscribeNotificationsToUser replaces the notifications subscribed by the user.
	SubscribeNotificationsToUser(ctx context.Context, userID string, notifications []string) error
	// UnsubscribeNotificationFromUser unsubscribes the user from a notification, if the unsubscribe key is valid.
	UnsubscribeNotificationFromUser(ctx context.Context, userID, notification, unsubscribeKey string) error
}

// TokensStorage is the storage of the auth tokens, provided by the application, such as the one of the storage plugin.
// The context is the one of the request.
type TokensStorage interface {
	// SaveAuthenticationToken saves an auth token and returns its ID, and whether the IP address is already known.
	SaveAuthenticationToken(ctx context.Context, details authentication.Details) (tokenID string, knownIPAddress bool, err error)
	// GetTokenExpirationDate returns the expiration date of an auth token.
	GetTokenExpirationDate(ctx context.Context, tokenID string) (time.Time, error)
	// DeleteTokenByID deletes an auth token.
	DeleteTokenByID(ctx context.Context, tokenID string) error
	// GetTokenDetailsByID returns the details of an auth token.
	GetTokenDetailsByID(ctx context.Context, tokenID string) (authentication.Token, error)
	// GetOwnedTokens returns the auth tokens of the user, to display them to the user.
	GetOwnedTokens(ctx context.Context, userID string) (any, error)
	// GetOwnedTokenByID returns an auth token, to display it to its owner, and the ID of its owner.
	GetOwnedTokenByID(ctx context.Context, tokenID string) (token any, ownerID string, err error)
	// UpdateTokenByID updates the name and the notifications of an auth token.
	UpdateTokenByID(ctx context.Context, tokenID, name string, enableNotifications bool, notifications []string) error
}

// ProcessesStorage is the storage of the processes bound to a user, such as the forgot password ones or the ones
// confirming the update of the account, provided by the application, such as the one of the storage plugin.
// The context is the one of the request.
type ProcessesStorage interface {
	// NewForgotPasswordProcessForUser creates a forgot password process for the user, replacing the existing one, and
	// returns its ID. The private key is the code to give to the user.
	NewForgotPasswordProcessForUser(ctx context.Context, userID, processPrivateKey string) (processID string, err error)
	// GetForgotPasswordProcessDetailsByID returns the ID of the user and the private key of a forgot password process.
	GetForgotPasswordProcessDetailsByID(ctx context.Context, processID string) (userID, processPrivateKey string, err error)
	// WrongAccessToForgotPasswordProcess records a wrong access to a forgot password process, to avoid brute force:
	// it increments the tries or deletes the process.
	WrongAccessToForgotPasswordProcess(ctx context.Context, processID string)
	// DeleteForgotPasswordProcess deletes a forgot password process.
	DeleteForgotPasswordProcess(ctx context.Context, processID string) error
	// CreateNewDeleteAccountProcess starts a process to delete the account of the user, replacing the existing one.
	CreateNewDeleteAccountProcess(ctx context.Context, userID string) (processID, processKey string, err error)
	// CreateNewDisableAccountProcess starts a process to disable the account of the user, replacing the existing one.
	CreateNewDisableAccountProcess(ctx context.Context, userID string) (processID, processKey string, err error)
	// CreateNewUpdatePasswordProcess starts a process to update the password of the user to the encrypted one,
	// replacing the existing one.
	CreateNewUpdatePasswordProcess(ctx context.Context, userID, encryptedPassword string) (processID, processKey string, err error)
	// CreateNewUpdateEmailAddressProcess starts a process to update the email address of the user, replacing the
	// existing one.
	CreateNewUpdateEmailAddressProcess(ctx context.Context, userID, emailAddress string) (processID, processKey string, err error)
	// CreateNewVerifyEmailAddressProcess starts a process to verify the email address of the user, replacing the
	// existing one.
	CreateNewVerifyEmailAddressProcess(ctx context.Context, userID string) (processID, processKey string, err error)
	// VerifyProcess verifies the key of a process and returns the ID of its user and its value. It returns a "not
	// found" error if the process does not exist or if the key is wrong.
	VerifyProcess(ctx context.Context, processID, key string) (userID, newValue string, err error)
	// DeleteProcessByID deletes a process.
	DeleteProcessByID(ctx context.Context, processID string) error
}

// AccountsStorage is the storage of the accounts of the users, provided by the application, such as the one of the
// storage plugin. The context is the one of the request.
type AccountsStorage interface {
	// GetUserDetailsByIdentifier returns the details of a user by its identifier: the username or the email address.
	GetUserDetailsByIdentifier(ctx context.Context, identifier string) (authentication.User, error)
	// GetUserDetailsByID returns the details of a user by its ID.
	GetUserDetailsByID(ctx context.Context, userID string) (authentication.User, error)
	// GetUserDetailsByFacebookUserID returns the details of a user by its Facebook ID.
	GetUserDetailsByFacebookUserID(ctx context.Context, facebookUserID string) (authentication.User, error)
	// GetUserDetailsByGoogleUserID returns the details of a user by its Google ID.
	GetUserDetailsByGoogleUserID(ctx context.Context, googleUserID string) (authentication.User, error)
	// GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser returns the details and the encrypted password of
	// an enabled user, by its identifier: the username or the email address.
	GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(ctx context.Context, identifier string) (authentication.User, string, error)
	// UpdateEncryptedPasswordByUserID replaces the encrypted password of a user, such as by the same password encrypted
	// with stronger parameters.
	UpdateEncryptedPasswordByUserID(ctx context.Context, userID, encryptedPassword string) error

	// UpdateUserPassword updates the password of a user.
	UpdateUserPassword(ctx context.Context, userID, password string, isPasswordEncrypted bool) error
	// UpdateUserEmailAddress updates the email address of a user.
	UpdateUserEmailAddress(ctx context.Context, userID, emailAddress string) error
	// UserEmailAddressIsVerified flags the email address of a user as valid.
	UserEmailAddressIsVerified(ctx context.Context, userID string) error
	// UpdateUserSocialNetworks updates the social network accounts of a user.
	UpdateUserSocialNetworks(ctx context.Context, userID, facebook, twitter, instagram string) error

	// DoesAccountWithUsernameExist returns true if an account with this username does exist already.
	DoesAccountWithUsernameExist(ctx context.Context, username string) (bool, error)
	// DoesAccountWithEmailAddressExist returns true if an account with this email address does exist already.
	DoesAccountWithEmailAddressExist(ctx context.Context, emailAddress string) (bool, error)
	// DoesUserWithFacebookUserIDExist returns true if a user with this Facebook ID is registered already.
	DoesUserWithFacebookUserIDExist(ctx context.Context, facebookUserID string) (bool, error)
	// DoesUserWithGoogleUserIDExist returns true if a user with this Google ID is registered already.
	DoesUserWithGoogleUserIDExist(ctx context.Context, googleUserID string) (bool, error)

	// RegisterByEmailAddress saves a newly registered user.
	RegisterByEmailAddress(ctx context.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error
	// RegisterByThirdPartyToken saves a newly registered user.
	RegisterByThirdPartyToken(ctx context.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error
	// ValidateUserRegistration validates the registration of a user.
	ValidateUserRegistration(ctx context.Context, userID, registrationKey string) error
	// CancelUserRegistration cancels the registration of a user.
	CancelUserRegistration(ctx context.Context, userID, registrationKey string) error
	// GetRegistrationDetails returns the details of the registration of a user.
	GetRegistrationDetails(ctx context.Context, userID string) (*authentication.Details, error)
	// SaveFacebookUserTokenDetailsToUser saves the details of the token of Facebook to the account of a user.
	SaveFacebookUserTokenDetailsToUser(ctx context.Context, userID string, tokenInfos validator.TokenInfos) error
	// SaveGoogleUserTokenDetailsToUser saves the details of the token of Google to the account of a user.
	SaveGoogleUserTokenDetailsToUser(ctx context.Context, userID string, tokenInfos validator.TokenInfos) error

	// GetOwnedAccountDetails returns the details of the account of a user, to display them to its owner.
	GetOwnedAccountDetails(ctx context.Context, userID string) (any, error)
	// UpdatedAccountInfos updates the first name and the last name of a user.
	UpdatedAccountInfos(ctx context.Context, userID, firstName, lastName string) error
	// UpdatedAccountSettings updates the locale and the timezone of a user.
	UpdatedAccountSettings(ctx context.Context, userID, locale, timezone string) error
	// DeleteUserByID deletes a user.
	DeleteUserByID(ctx context.Context, userID string) error
	// DisableUserByID disables a user and returns its details.
	DisableUserByID(ctx context.Context, userID string) (authentication.User, error)
	// EnableUserByID enables a disabled user and returns its details.
	EnableUserByID(ctx context.Context, userID string) (authentication.User, error)
}
