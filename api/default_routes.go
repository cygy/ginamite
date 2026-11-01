package api

import (
	"fmt"
	"reflect"

	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/common/log"
)

// requirement is a default route and the functions of the application its handler calls.
type requirement struct {
	route     *route.PathProperties
	functions []any
}

// requirements returns, for each default route which calls the functions of the application, the functions it needs.
// The routes without any, such as the status or the locales, are always available.
func (s *Server) requirements(routes *route.DefaultRoutes) []requirement {
	f := &s.Functions
	account, authentication, registration := &routes.Account.Paths, &routes.Authentication.Paths, &routes.Registration.Paths
	forgotPassword, admin := &routes.ForgotPassword.Paths, &routes.Admin.Paths

	return []requirement{
		{&routes.Username.Paths.Availability, []any{f.DoesAccountWithUsernameExist}},
		{&routes.Terms.Paths.Accept, []any{f.UpdateVersionOfTermsAcceptedByUser}},
		{&routes.Contact.Paths.Post, []any{f.CreateContact}},

		{&registration.RegisterWithPassword, []any{f.RegisterByEmailAddress, f.DoesAccountWithEmailAddressExist, f.DoesAccountWithUsernameExist}},
		{&registration.RegisterWithFacebook, []any{f.RegisterByThirdPartyToken, f.DoesUserWithFacebookUserIDExist}},
		{&registration.RegisterWithGoogle, []any{f.RegisterByThirdPartyToken, f.DoesUserWithGoogleUserIDExist}},
		{&registration.Validate, []any{f.ValidateUserRegistration, f.GetRegistrationDetails}},
		{&registration.Cancel, []any{f.CancelUserRegistration, f.GetRegistrationDetails}},

		{&forgotPassword.ForgotPassword, []any{f.GetUserDetailsByIdentifier, f.NewForgotPasswordProcessForUser}},
		{&forgotPassword.Validate, []any{f.GetForgotPasswordProcessDetailsByID, f.UpdateUserPassword}},
		{&forgotPassword.Cancel, []any{f.GetForgotPasswordProcessDetailsByID, f.DeleteForgotPasswordProcess}},

		{&authentication.LoginWithPassword, []any{f.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser, f.SaveAuthenticationToken}},
		{&authentication.LoginWithFacebook, []any{f.GetUserDetailsByFacebookUserID, f.SaveAuthenticationToken}},
		{&authentication.LoginWithGoogle, []any{f.GetUserDetailsByGoogleUserID, f.SaveAuthenticationToken}},
		{&authentication.Logout, []any{f.DeleteTokenByID}},
		{&authentication.GetTokens, []any{f.GetOwnedTokens}},
		{&authentication.GetToken, []any{f.GetOwnedTokenByID}},
		{&authentication.UpdateToken, []any{f.GetTokenDetailsByID, f.UpdateTokenByID}},
		{&authentication.RefreshToken, []any{f.SaveAuthenticationToken}},
		{&authentication.DeleteToken, []any{f.GetTokenDetailsByID, f.DeleteTokenByID}},
		{&authentication.DeleteTokenWithKey, []any{f.GetTokenDetailsByID, f.DeleteTokenByID}},

		{&account.GetNotifications, []any{f.NotificationsReceivedByUser}},
		{&account.GetNotification, []any{f.NotificationReceivedByUser}},
		{&account.DeleteNotification, []any{f.DeleteUserNotification}},
		{&account.UpdateNotificationRead, []any{f.UpdateUserNotificationRead}},
		{&account.UpdateNotificationNotified, []any{f.UpdateUserNotificationNotified}},
		{&account.GetNotificationsSettings, []any{f.NotificationsSubscribedByUser}},
		{&account.SaveNotificationsSettings, []any{f.SubscribeNotificationsToUser}},
		{&account.UnsubscribeFromNotification, []any{f.UnsubscribeNotificationFromUser}},
		{&account.GetDetails, []any{f.GetOwnedAccountDetails}},
		{&account.UpdateDetails, []any{f.UpdatedAccountInfos}},
		{&account.UpdateSettings, []any{f.UpdatedAccountSettings}},
		{&account.UpdateSocialNetworks, []any{f.UpdateUserSocialNetworks}},
		{&account.UpdateEmailAddress, []any{f.CreateNewUpdateEmailAddressProcess, f.GetUserDetailsByID}},
		{&account.ValidateUpdateEmailAddress, []any{f.VerifyProcess, f.UpdateUserEmailAddress, f.DeleteProcessByID}},
		{&account.CancelUpdateEmailAddress, []any{f.VerifyProcess, f.DeleteProcessByID}},
		{&account.ValidateNewEmailAddress, []any{f.VerifyProcess, f.UserEmailAddressIsVerified, f.DeleteProcessByID}},
		{&account.ConfirmEmailAddress, []any{f.CreateNewVerifyEmailAddressProcess, f.GetUserDetailsByID}},
		{&account.UpdatePassword, []any{f.CreateNewUpdatePasswordProcess, f.GetUserDetailsByID}},
		{&account.ValidateUpdatePassword, []any{f.VerifyProcess, f.UpdateUserPassword, f.DeleteProcessByID}},
		{&account.CancelUpdatePassword, []any{f.VerifyProcess, f.DeleteProcessByID}},
		{&account.Delete, []any{f.CreateNewDeleteAccountProcess, f.GetUserDetailsByID}},
		{&account.ValidateDelete, []any{f.VerifyProcess, f.DeleteUserByID, f.DeleteProcessByID}},
		{&account.CancelDelete, []any{f.VerifyProcess, f.DeleteProcessByID}},
		{&account.Disable, []any{f.CreateNewDisableAccountProcess, f.GetUserDetailsByID}},
		{&account.ValidateDisable, []any{f.VerifyProcess, f.DisableUserByID, f.DeleteProcessByID}},
		{&account.CancelDisable, []any{f.VerifyProcess, f.DeleteProcessByID}},

		{&admin.GetEnabledUsers, []any{f.GetEnabledUsersForAdmin}},
		{&admin.GetDisabledUsers, []any{f.GetDisabledUsersForAdmin}},
		{&admin.UpdateUserPassword, []any{f.UpdatePasswordByAdmin}},
		{&admin.UpdateUserInformations, []any{f.UpdateInformationsByAdmin}},
		{&admin.UpdateUserDescription, []any{f.UpdateDescriptionByAdmin}},
		{&admin.UpdateUserSocial, []any{f.UpdateSocialByAdmin}},
		{&admin.UpdateUserSettings, []any{f.UpdateSettingsByAdmin}},
		{&admin.UpdateUserNotifications, []any{f.UpdateNotificationsByAdmin}},
		{&admin.GetUserTokens, []any{f.GetTokensForAdmin}},
		{&admin.DeleteUserToken, []any{f.DeleteTokenByIDByAdmin, f.GetTokenDetailsByID}},
		{&admin.SetAbilitiesToUser, []any{f.SetAbilitiesToUser}},
		{&admin.AddAbilitiesToUser, []any{f.AddAbilitiesToUser}},
		{&admin.RemoveAbilitiesFromUser, []any{f.RemoveAbilitiesFromUser}},
		{&admin.DeleteUser, []any{f.DeleteUserByID, f.GetUserDetailsByID}},
		{&admin.DisableUser, []any{f.DisableUserByID}},
		{&admin.EnableUser, []any{f.EnableUserByID}},
		{&admin.GetContacts, []any{f.GetContacts}},
		{&admin.GetContact, []any{f.GetContact}},
		{&admin.UpdateContact, []any{f.UpdateContact}},
	}
}

// disableUndefinedRoutes disables the default routes whose handler calls a function of the application which is
// undefined: an application only implements what it uses. The application can still enable a route when it configures
// the default routes.
func (s *Server) disableUndefinedRoutes(routes *route.DefaultRoutes) {
	disabled := 0

	for _, r := range s.requirements(routes) {
		if !r.route.Enabled || allDefined(r.functions) {
			continue
		}

		r.route.Enabled = false
		disabled++
	}

	if disabled > 0 {
		log.Debug(fmt.Sprintf("%d default routes are disabled: the functions they need are undefined", disabled))
	}

	if routes.Authentication.Paths.LoginWithPassword.Enabled && s.Functions.UpdateEncryptedPasswordByUserID == nil {
		log.Warn("The function 'UpdateEncryptedPasswordByUserID' is undefined: the weak password hashes are not upgraded.")
	}
}

// allDefined reports whether none of the functions is nil.
func allDefined(functions []any) bool {
	for _, function := range functions {
		if reflect.ValueOf(function).IsNil() {
			return false
		}
	}

	return true
}
