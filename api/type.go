package api

import (
	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/rights"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/api/route/contact"
	"github.com/cygy/ginamite/api/route/terms"
	"github.com/cygy/ginamite/api/route/user"
	"github.com/cygy/ginamite/common/authentication"
)

// Server : definition of an API server.
type Server struct {
	RoutesFile string
	Functions  Handlers

	customRoutes []route.ConfigureCustomRoutes
}

// Handlers : list of the functions of an API.
type Handlers struct {
	RecurringTasks                                                func()
	NewMessageFromQueueCache                                      func(message string) bool
	MiddlewareIsValidAuthToken                                    context.IsAuthTokenValidFunc
	MiddlewareGetUserAndLocaleFromAuthToken                       context.GetUserAndLocaleFromAuthTokenFunc
	MiddlewareGetUserAbilities                                    rights.GetUserAbilitesFunc
	MiddlewareGetLatestVersionOfTermsAcceptedByUser               context.GetLatestVersionOfTermsAcceptedByUserFunc
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
