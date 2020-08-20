package account

var (
	// NotificationsReceivedByUser : function to get all the notifications received by the user.
	NotificationsReceivedByUser NotificationsReceivedByUserFunc

	// NotificationReceivedByUser : function to get a notification received by the user.
	NotificationReceivedByUser NotificationReceivedByUserFunc

	// UpdateUserNotificationRead : function to update the read flag of a notification received by the user.
	UpdateUserNotificationRead UpdateUserNotificationReadFunc

	// UpdateUserNotificationNotified : function to update the notified flag of a notification received by the user.
	UpdateUserNotificationNotified UpdateUserNotificationNotifiedFunc

	// DeleteUserNotification : function to delete a notification received by the user.
	DeleteUserNotification DeleteUserNotificationFunc

	// NotificationsSubscribedByUser : function to get all the notifications subscribed by the user.
	NotificationsSubscribedByUser NotificationsSubscribedByUserFunc

	// SubscribeNotificationsToUser : function to subscribe some notifications to the user.
	SubscribeNotificationsToUser SubscribeNotificationsToUserFunc

	// UnsubscribeNotificationFromUser : function to unsubscribe a notification from a user.
	UnsubscribeNotificationFromUser UnsubscribeNotificationFromUserFunc

	// GetUserDetailsByIdentifier : function to get the user details from its identifier (username, email addresse, what you want.)
	GetUserDetailsByIdentifier GetUserDetailsByIdentifierFunc

	// GetUserDetailsByID : function to get the user details from its unique identifier.
	GetUserDetailsByID GetUserDetailsByIDFunc

	// GetUserDetailsByFacebookUserID : function to get the user details from its unique facebook identifier.
	GetUserDetailsByFacebookUserID GetUserDetailsBySocialNetworkIDFunc

	// GetUserDetailsByGoogleUserID : function to get the user details from its unique facebook identifier.
	GetUserDetailsByGoogleUserID GetUserDetailsBySocialNetworkIDFunc

	// GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser : function to get the user details from its identifier (username, email addresse, what you want.)
	GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUserFunc

	// NewForgotPasswordProcessForUser : function to create a new forgot password process for a user. Return the unique identifier of the process.
	NewForgotPasswordProcessForUser NewForgotPasswordProcessForUserFunc

	// GetForgotPasswordProcessDetailsByID : function to get the forgot password process details from its identifier.
	GetForgotPasswordProcessDetailsByID GetForgotPasswordProcessDetailsByIDFunc

	// WrongAccessToForgotPasswordProcess : function to avoid brute force, increment tries or delete the process.
	WrongAccessToForgotPasswordProcess WrongAccessToForgotPasswordProcessFunc

	// UpdateUserPassword : function to update the password of an user.
	UpdateUserPassword UpdateUserPasswordFunc

	// UpdateUserEmailAddress : function to update the email address of an user.
	UpdateUserEmailAddress UpdateUserEmailAddressFunc

	// UserEmailAddressIsVerified : function to flag the email address of a user as valid.
	UserEmailAddressIsVerified UserEmailAddressIsVerifiedFunc

	// UpdateUserSocialNetworks : updates the social network accounts of the user.
	UpdateUserSocialNetworks UpdateUserSocialNetworksFunc

	// DeleteForgotPasswordProcess : function to delete a forgot password process by ID.
	DeleteForgotPasswordProcess DeleteForgotPasswordProcessFunc

	// ValidateUserRegistration : function to validate the registration of a user by its ID.
	ValidateUserRegistration ValidateUserRegistrationFunc

	// CancelUserRegistration : function to cancel the registration of a user by its ID.
	CancelUserRegistration CancelUserRegistrationFunc

	// GetRegistrationDetails : function to get the details of the registration of a user.
	GetRegistrationDetails GetRegistrationDetailsFunc

	// SaveAuthenticationToken : function to save the authentication details and return the unique ID.
	SaveAuthenticationToken SaveAuthenticationTokenFunc

	// DoesAccountWithUsernameExist : returns true if an account with this username does exist already.
	DoesAccountWithUsernameExist DoesAccountWithUsernameExistFunc

	// DoesAccountWithEmailAddressExist : returns true if an account with this email address does exist already.
	DoesAccountWithEmailAddressExist DoesAccountWithEmailAddressExistFunc

	// RegisterByEmailAddress : saves a newly registered user.
	RegisterByEmailAddress RegisterByEmailAddressFunc

	// RegisterByThirdPartyToken : saves a newly registered user.
	RegisterByThirdPartyToken RegisterByThirdPartyTokenFunc

	// DoesUserWithFacebookUserIDExist : return true if a user with a facebook user id is registered already.
	DoesUserWithFacebookUserIDExist DoesUserWithSocialNetworkUserIDExistFunc

	// DoesUserWithGoogleUserIDExist : return true if a user with a google user id is registered already.
	DoesUserWithGoogleUserIDExist DoesUserWithSocialNetworkUserIDExistFunc

	// SaveFacebookUserTokenDetailsToUser : function to save the auth token details from facebook to an account.
	SaveFacebookUserTokenDetailsToUser SaveSocialNetworkUserTokenDetailsToUserFunc

	// SaveGoogleUserTokenDetailsToUser : function to save the auth token details from google to an account.
	SaveGoogleUserTokenDetailsToUser SaveSocialNetworkUserTokenDetailsToUserFunc

	// GetTokenExpirationDate : function to get the expiration date of a token from its unique identifier.
	GetTokenExpirationDate GetTokenExpirationDateFunc

	// DeleteTokenByID : delete an auth token by its ID.
	DeleteTokenByID DeleteTokenByIDFunc

	// GetTokenDetailsByID : function to get the token details from its unique identifier.
	GetTokenDetailsByID GetTokenDetailsByIDFunc

	// GetOwnedTokens : function to get the auth tokens of a user.
	GetOwnedTokens GetOwnedTokensFunc

	// GetOwnedTokenByID : function to get the auth token by its ID and the ID of the token's owner.
	GetOwnedTokenByID GetOwnedTokenByIDFunc

	// UpdateTokenByID : function to update the properties of a token.
	UpdateTokenByID UpdateTokenByIDFunc

	// GetOwnedAccountDetails : function to get the user details from its unique identifier.
	GetOwnedAccountDetails GetOwnedAccountDetailsFunc

	// UpdatedAccountInfos : function to update the user infos.
	UpdatedAccountInfos UpdatedAccountInfosFunc

	// UpdatedAccountSettings : function to update the user settings.
	UpdatedAccountSettings UpdatedAccountSettingsFunc

	// CreateNewDeleteAccountProcess : function to start a new process to delete an account.
	CreateNewDeleteAccountProcess CreateNewProcessFunc

	// CreateNewDisableAccountProcess : function to start a new process to disable an account.
	CreateNewDisableAccountProcess CreateNewProcessFunc

	// CreateNewUpdatePasswordProcess : function to start a new process to update the password of an account.
	CreateNewUpdatePasswordProcess CreateNewProcessWithValueFunc

	// CreateNewUpdateEmailAddressProcess : function to start a new process to update the email address of an account.
	CreateNewUpdateEmailAddressProcess CreateNewProcessWithValueFunc

	// CreateNewVerifyEmailAddressProcess : function to start a new process to verify the email address of an account.
	CreateNewVerifyEmailAddressProcess CreateNewProcessFunc

	// VerifyProcess : function to verify a process and get the ID of the bound user.
	VerifyProcess VerifyProcessFunc

	// DeleteProcessByID : deletes a process by its ID.
	DeleteProcessByID DeleteProcessByIDFunc

	// DeleteUserByID : deletes a user by its ID.
	DeleteUserByID DeleteUserByIDFunc

	// DisableUserByID : disables a user by its ID.
	DisableUserByID DisableUserByIDFunc

	// EnableUserByID : enables a user by its ID.
	EnableUserByID EnableUserByIDFunc
)
