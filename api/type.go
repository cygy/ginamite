package api

import (
	tokencache "github.com/cygy/ginamite/api/cache/authtoken"
	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	"github.com/cygy/ginamite/api/rights"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/api/route/contact"
	"github.com/cygy/ginamite/api/route/terms"
	"github.com/cygy/ginamite/api/route/user"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

// Server defines an API server.
type Server struct {
	// Config is the configuration of the application, set by the main server before it starts.
	Config *config.Configuration
	// Queue sends the messages to the other servers, set by the main server before it starts.
	Queue *queue.Queue
	// Localization translates the messages of the responses, set by the main server before it starts.
	Localization *localization.Engine
	// Timezones are the timezones of the countries, set by the main server before it starts.
	Timezones *timezone.Configuration
	// Notifications is the registry of the notifications the users can subscribe to, set by the main server before it
	// starts.
	Notifications *notifications.Registry
	Functions     Handlers

	// RevokedTokens are the auth tokens revoked, which are rejected until they expire. The cache messages of the queue
	// add tokens to it.
	RevokedTokens *tokencache.Revoked

	// Middlewares are added to the Gin engine, in order, before the middleware that sets the locale and the routes.
	// A database plugin uses them to put its connection in the context of the requests.
	Middlewares []gin.HandlerFunc

	customRoutes []route.ConfigureCustomRoutes
}

// Handlers lists the functions of an API.
type Handlers struct {
	RecurringTasks                                                func()
	IsValidID                                                     func(id string) bool
	NewMessageFromQueueCache                                      func(message string) bool
	MiddlewareIsValidAuthToken                                    ctxAuthToken.IsValidFunc
	MiddlewareGetUserAndLocaleFromAuthToken                       ctxAuthToken.GetUserAndLocaleFunc
	MiddlewareGetUserAbilities                                    rights.GetUserAbilitiesFunc
	MiddlewareGetLatestVersionOfTermsAcceptedByUser               terms.GetLatestVersionOfTermsAcceptedByUserFunc
	BuildTokenFromID                                              authentication.BuildTokenFunc
	ExtraPropertiesForTokenWithID                                 authentication.ExtraPropertiesForTokenWithIDFunc
	ExtendTokenExpirationDateFromID                               authentication.ExtendTokenExpirationDateFunc
	RouterConfigureDefaultRoutes                                  route.ConfigureDefaultRoutes
	UpdateVersionOfTermsAcceptedByUser                            terms.UpdateVersionOfTermsAcceptedByUserFunc
	GetEnabledUsersForAdmin                                       user.GetUsersForAdminFunc
	GetDisabledUsersForAdmin                                      user.GetUsersForAdminFunc
	UpdatePasswordByAdmin                                         user.UpdatePasswordByAdminFunc
	UpdateInformationsByAdmin                                     user.UpdateInformationsByAdminFunc
	UpdateDescriptionByAdmin                                      user.UpdateDescriptionByAdminFunc
	UpdateSocialByAdmin                                           user.UpdateSocialByAdminFunc
	UpdateSettingsByAdmin                                         user.UpdateSettingsByAdminFunc
	UpdateNotificationsByAdmin                                    user.UpdateNotificationsByAdminFunc
	GetTokensForAdmin                                             user.GetTokensForAdminFunc
	DeleteTokenByIDByAdmin                                        user.DeleteTokenByIDByAdminFunc
	SetAbilitiesToUser                                            user.SetAbilitiesToUserFunc
	AddAbilitiesToUser                                            user.AddAbilitiesToUserFunc
	RemoveAbilitiesFromUser                                       user.RemoveAbilitiesFromUserFunc
	NotificationsReceivedByUser                                   account.NotificationsReceivedByUserFunc
	NotificationReceivedByUser                                    account.NotificationReceivedByUserFunc
	UpdateUserNotificationRead                                    account.UpdateUserNotificationReadFunc
	UpdateUserNotificationNotified                                account.UpdateUserNotificationNotifiedFunc
	DeleteUserNotification                                        account.DeleteUserNotificationFunc
	NotificationsSubscribedByUser                                 account.NotificationsSubscribedByUserFunc
	SubscribeNotificationsToUser                                  account.SubscribeNotificationsToUserFunc
	UnsubscribeNotificationFromUser                               account.UnsubscribeNotificationFromUserFunc
	GetUserDetailsByIdentifier                                    account.GetUserDetailsByIdentifierFunc
	GetUserDetailsByID                                            account.GetUserDetailsByIDFunc
	NewForgotPasswordProcessForUser                               account.NewForgotPasswordProcessForUserFunc
	GetForgotPasswordProcessDetailsByID                           account.GetForgotPasswordProcessDetailsByIDFunc
	WrongAccessToForgotPasswordProcess                            account.WrongAccessToForgotPasswordProcessFunc
	UpdateUserPassword                                            account.UpdateUserPasswordFunc
	UpdateUserEmailAddress                                        account.UpdateUserEmailAddressFunc
	UserEmailAddressIsVerified                                    account.UserEmailAddressIsVerifiedFunc
	UpdateUserSocialNetworks                                      account.UpdateUserSocialNetworksFunc
	DeleteForgotPasswordProcess                                   account.DeleteForgotPasswordProcessFunc
	ValidateUserRegistration                                      account.ValidateUserRegistrationFunc
	CancelUserRegistration                                        account.CancelUserRegistrationFunc
	GetRegistrationDetails                                        account.GetRegistrationDetailsFunc
	SaveAuthenticationToken                                       account.SaveAuthenticationTokenFunc
	DoesAccountWithUsernameExist                                  account.DoesAccountWithUsernameExistFunc
	DoesAccountWithEmailAddressExist                              account.DoesAccountWithEmailAddressExistFunc
	RegisterByEmailAddress                                        account.RegisterByEmailAddressFunc
	RegisterByThirdPartyToken                                     account.RegisterByThirdPartyTokenFunc
	DoesUserWithFacebookUserIDExist                               account.DoesUserWithSocialNetworkUserIDExistFunc
	DoesUserWithGoogleUserIDExist                                 account.DoesUserWithSocialNetworkUserIDExistFunc
	GetUserDetailsByFacebookUserID                                account.GetUserDetailsBySocialNetworkIDFunc
	GetUserDetailsByGoogleUserID                                  account.GetUserDetailsBySocialNetworkIDFunc
	SaveFacebookUserTokenDetailsToUser                            account.SaveSocialNetworkUserTokenDetailsToUserFunc
	SaveGoogleUserTokenDetailsToUser                              account.SaveSocialNetworkUserTokenDetailsToUserFunc
	GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser account.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUserFunc
	UpdateEncryptedPasswordByUserID                               account.UpdateEncryptedPasswordByUserIDFunc
	GetTokenExpirationDate                                        account.GetTokenExpirationDateFunc
	DeleteTokenByID                                               account.DeleteTokenByIDFunc
	GetTokenDetailsByID                                           account.GetTokenDetailsByIDFunc
	GetOwnedTokens                                                account.GetOwnedTokensFunc
	GetOwnedTokenByID                                             account.GetOwnedTokenByIDFunc
	UpdateTokenByID                                               account.UpdateTokenByIDFunc
	GetOwnedAccountDetails                                        account.GetOwnedAccountDetailsFunc
	UpdatedAccountInfos                                           account.UpdatedAccountInfosFunc
	UpdatedAccountSettings                                        account.UpdatedAccountSettingsFunc
	CreateNewDeleteAccountProcess                                 account.CreateNewProcessFunc
	CreateNewDisableAccountProcess                                account.CreateNewProcessFunc
	CreateNewUpdatePasswordProcess                                account.CreateNewProcessWithValueFunc
	CreateNewUpdateEmailAddressProcess                            account.CreateNewProcessWithValueFunc
	CreateNewVerifyEmailAddressProcess                            account.CreateNewProcessFunc
	VerifyProcess                                                 account.VerifyProcessFunc
	DeleteProcessByID                                             account.DeleteProcessByIDFunc
	DeleteUserByID                                                account.DeleteUserByIDFunc
	DisableUserByID                                               account.DisableUserByIDFunc
	EnableUserByID                                                account.EnableUserByIDFunc
	CreateContact                                                 contact.CreateContactFunc
	UpdateContact                                                 contact.UpdateContactFunc
	GetContacts                                                   contact.GetContactsFunc
	GetContact                                                    contact.GetContactFunc
}
