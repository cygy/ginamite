package api

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/cygy/ginamite/api/route"
)

// The features of the API: the sets of default routes the application chooses with the setting api_features. A feature
// holds its own routes of administration, and needs the functions of the application they call.
const (
	// FeatureStatus: the status of the API.
	FeatureStatus = "status"
	// FeatureLocales: the locales supported by the API.
	FeatureLocales = "locales"
	// FeatureTimezones: the timezones of the countries.
	FeatureTimezones = "timezones"
	// FeatureTerms: the terms of use, their version, and their acceptance by the users.
	FeatureTerms = "terms"
	// FeatureNotifications: the notifications of the users, their center and their settings.
	FeatureNotifications = "notifications"
	// FeatureRegistration: the registration with a password, and the availability of the usernames.
	FeatureRegistration = "registration"
	// FeatureForgotPassword: the reset of a forgotten password.
	FeatureForgotPassword = "forgot_password"
	// FeatureAuth: the login with a password, the auth tokens of the users, and their administration.
	FeatureAuth = "auth"
	// FeatureFacebook: the registration and the login with Facebook.
	FeatureFacebook = "facebook"
	// FeatureGoogle: the registration and the login with Google.
	FeatureGoogle = "google"
	// FeatureAccount: the account of the users and, for the administrators, the management of the users.
	FeatureAccount = "account"
	// FeatureContact: the contact messages and, for the administrators, their management.
	FeatureContact = "contact"
)

// requirement is a default route and the names of the functions of the application, see Server.Account and the other groups of functions, its handler calls.
type requirement struct {
	route     *route.PathProperties
	functions []string
}

// feature is a set of default routes.
type feature struct {
	name     string
	routes   []requirement
	needsJWT bool // The routes sign or verify auth tokens, which needs the setting jwt.secret.
}

// Features returns the names of the features of the API.
func Features() []string {
	var routes route.DefaultRoutes

	features := (&Server{}).features(&routes)
	names := make([]string, len(features))
	for i, f := range features {
		names[i] = f.name
	}

	return names
}

// features returns the features of the API, with the routes of the default routes.
func (s *Server) features(routes *route.DefaultRoutes) []feature {
	account, authentication, registration := &routes.Account.Paths, &routes.Authentication.Paths, &routes.Registration.Paths
	forgotPassword, admin, terms := &routes.ForgotPassword.Paths, &routes.Admin.Paths, &routes.Terms.Paths

	return []feature{
		{FeatureStatus, []requirement{{&routes.Status.Paths.Get, nil}}, false},
		{FeatureLocales, []requirement{{&routes.Locales.Paths.Get, nil}}, false},
		{FeatureTimezones, []requirement{{&routes.Timezones.Paths.Get, nil}}, false},
		{FeatureTerms, []requirement{
			{&terms.GetTerms, nil},
			{&terms.GetVersion, nil},
			{&terms.Accept, []string{"Terms"}},
		}, true},
		{FeatureNotifications, []requirement{
			{&routes.Notifications.Paths.Get, nil},
			{&account.GetNotifications, []string{"Notifications"}},
			{&account.GetNotification, []string{"Notifications"}},
			{&account.DeleteNotification, []string{"Notifications"}},
			{&account.UpdateNotificationRead, []string{"Notifications"}},
			{&account.UpdateNotificationNotified, []string{"Notifications"}},
			{&account.GetNotificationsSettings, []string{"Notifications"}},
			{&account.SaveNotificationsSettings, []string{"Notifications"}},
			{&account.UnsubscribeFromNotification, []string{"Notifications"}},
			{&admin.UpdateUserNotifications, []string{"Users"}},
		}, true},
		{FeatureRegistration, []requirement{
			{&routes.Username.Paths.Availability, []string{"DoesAccountWithUsernameExist"}},
			{&registration.RegisterWithPassword, []string{"RegisterByEmailAddress", "DoesAccountWithEmailAddressExist", "DoesAccountWithUsernameExist"}},
			{&registration.Validate, []string{"ValidateUserRegistration", "GetRegistrationDetails"}},
			{&registration.Cancel, []string{"CancelUserRegistration", "GetRegistrationDetails"}},
		}, true},
		{FeatureForgotPassword, []requirement{
			{&forgotPassword.ForgotPassword, []string{"GetUserDetailsByIdentifier", "NewForgotPasswordProcessForUser"}},
			{&forgotPassword.Validate, []string{"GetForgotPasswordProcessDetailsByID", "UpdateUserPassword"}},
			{&forgotPassword.Cancel, []string{"GetForgotPasswordProcessDetailsByID", "DeleteForgotPasswordProcess"}},
		}, true},
		{FeatureAuth, []requirement{
			{&authentication.LoginWithPassword, []string{"GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser", "SaveAuthenticationToken"}},
			{&authentication.Logout, []string{"DeleteTokenByID"}},
			{&authentication.GetTokens, []string{"GetOwnedTokens"}},
			{&authentication.GetToken, []string{"GetOwnedTokenByID"}},
			{&authentication.UpdateToken, []string{"GetTokenDetailsByID", "UpdateTokenByID"}},
			{&authentication.RefreshToken, []string{"SaveAuthenticationToken"}},
			{&authentication.DeleteToken, []string{"GetTokenDetailsByID", "DeleteTokenByID"}},
			{&authentication.DeleteTokenWithKey, []string{"GetTokenDetailsByID", "DeleteTokenByID"}},
			{&admin.GetUserTokens, []string{"Users"}},
			{&admin.DeleteUserToken, []string{"Users", "GetTokenDetailsByID"}},
		}, true},
		{FeatureFacebook, []requirement{
			{&registration.RegisterWithFacebook, []string{"RegisterByThirdPartyToken", "DoesUserWithFacebookUserIDExist"}},
			{&authentication.LoginWithFacebook, []string{"GetUserDetailsByFacebookUserID", "SaveAuthenticationToken"}},
		}, true},
		{FeatureGoogle, []requirement{
			{&registration.RegisterWithGoogle, []string{"RegisterByThirdPartyToken", "DoesUserWithGoogleUserIDExist"}},
			{&authentication.LoginWithGoogle, []string{"GetUserDetailsByGoogleUserID", "SaveAuthenticationToken"}},
		}, true},
		{FeatureAccount, []requirement{
			{&account.GetDetails, []string{"GetOwnedAccountDetails"}},
			{&account.UpdateDetails, []string{"UpdatedAccountInfos"}},
			{&account.UpdateSettings, []string{"UpdatedAccountSettings"}},
			{&account.UpdateSocialNetworks, []string{"UpdateUserSocialNetworks"}},
			{&account.UpdateEmailAddress, []string{"CreateNewUpdateEmailAddressProcess", "GetUserDetailsByID"}},
			{&account.ValidateUpdateEmailAddress, []string{"VerifyProcess", "UpdateUserEmailAddress", "DeleteProcessByID"}},
			{&account.CancelUpdateEmailAddress, []string{"VerifyProcess", "DeleteProcessByID"}},
			{&account.ValidateNewEmailAddress, []string{"VerifyProcess", "UserEmailAddressIsVerified", "DeleteProcessByID"}},
			{&account.ConfirmEmailAddress, []string{"CreateNewVerifyEmailAddressProcess", "GetUserDetailsByID"}},
			{&account.UpdatePassword, []string{"CreateNewUpdatePasswordProcess", "GetUserDetailsByID"}},
			{&account.ValidateUpdatePassword, []string{"VerifyProcess", "UpdateUserPassword", "DeleteProcessByID"}},
			{&account.CancelUpdatePassword, []string{"VerifyProcess", "DeleteProcessByID"}},
			{&account.Delete, []string{"CreateNewDeleteAccountProcess", "GetUserDetailsByID"}},
			{&account.ValidateDelete, []string{"VerifyProcess", "DeleteUserByID", "DeleteProcessByID"}},
			{&account.CancelDelete, []string{"VerifyProcess", "DeleteProcessByID"}},
			{&account.Disable, []string{"CreateNewDisableAccountProcess", "GetUserDetailsByID"}},
			{&account.ValidateDisable, []string{"VerifyProcess", "DisableUserByID", "DeleteProcessByID"}},
			{&account.CancelDisable, []string{"VerifyProcess", "DeleteProcessByID"}},
			{&admin.GetEnabledUsers, []string{"Users"}},
			{&admin.GetDisabledUsers, []string{"Users"}},
			{&admin.UpdateUserPassword, []string{"Users"}},
			{&admin.UpdateUserInformations, []string{"Users"}},
			{&admin.UpdateUserDescription, []string{"Users"}},
			{&admin.UpdateUserSocial, []string{"Users"}},
			{&admin.UpdateUserSettings, []string{"Users"}},
			{&admin.SetAbilitiesToUser, []string{"Users"}},
			{&admin.AddAbilitiesToUser, []string{"Users"}},
			{&admin.RemoveAbilitiesFromUser, []string{"Users"}},
			{&admin.DeleteUser, []string{"DeleteUserByID", "GetUserDetailsByID"}},
			{&admin.DisableUser, []string{"DisableUserByID"}},
			{&admin.EnableUser, []string{"EnableUserByID"}},
		}, true},
		{FeatureContact, []requirement{
			{&routes.Contact.Paths.Post, []string{"Contacts"}},
			{&admin.GetContacts, []string{"Contacts"}},
			{&admin.GetContact, []string{"Contacts"}},
			{&admin.UpdateContact, []string{"Contacts"}},
		}, true},
	}
}

// selectRoutes loads, among the default routes, the ones of the enabled features only. It returns an error naming the
// unknown features, or the functions of the application the features need and which are undefined. It reports whether
// the enabled features need the setting jwt.secret.
func (s *Server) selectRoutes(routes *route.DefaultRoutes, enabled []string) (needsJWT bool, err error) {
	features := s.features(routes)

	var unknown []string
	for _, name := range enabled {
		if !slices.ContainsFunc(features, func(f feature) bool { return f.name == name }) {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		names := make([]string, len(features))
		for i, f := range features {
			names[i] = f.name
		}

		return false, fmt.Errorf("unknown features %v in the setting api_features, the features are: %s", unknown, strings.Join(names, ", "))
	}

	groups := []reflect.Value{reflect.ValueOf(s.Account)}
	missing := map[string]bool{}
	for _, f := range features {
		isEnabled := slices.Contains(enabled, f.name)
		needsJWT = needsJWT || (isEnabled && f.needsJWT)

		for _, r := range f.routes {
			r.route.Enabled = isEnabled
			if !isEnabled {
				continue
			}

			for _, function := range r.functions {
				if s.isUndefined(groups, function) {
					missing[function] = true
				}
			}
		}
	}

	if len(missing) > 0 {
		names := make([]string, 0, len(missing))
		for name := range missing {
			names = append(names, name)
		}
		sort.Strings(names)

		return needsJWT, fmt.Errorf("the features of the setting api_features need functions or storages which are undefined: %s", strings.Join(names, ", "))
	}

	return needsJWT, nil
}

// isUndefined reports whether the storage, or the function, with the name is undefined. The names of the storages are
// the fields of the server, the names of the functions are the fields of the groups of functions.
func (s *Server) isUndefined(groups []reflect.Value, name string) bool {
	switch name {
	case "Contacts":
		return s.Contacts == nil
	case "Terms":
		return s.Terms == nil
	case "Users":
		return s.Users == nil
	case "Notifications":
		return s.Account.Notifications == nil
	}

	for _, group := range groups {
		if field := group.FieldByName(name); field.IsValid() {
			return field.IsNil()
		}
	}

	panic("api: the feature needs the unknown function " + name)
}
