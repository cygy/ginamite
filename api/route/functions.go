package route

import (
	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/middleware"
	"github.com/cygy/ginamite/api/rights"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/api/route/contact"
	"github.com/cygy/ginamite/api/route/debug"
	"github.com/cygy/ginamite/api/route/locale"
	"github.com/cygy/ginamite/api/route/notifications"
	"github.com/cygy/ginamite/api/route/status"
	"github.com/cygy/ginamite/api/route/terms"
	"github.com/cygy/ginamite/api/route/timezone"
	"github.com/cygy/ginamite/api/route/user"
	notificationsSettings "github.com/cygy/ginamite/common/notifications"

	"github.com/gin-gonic/gin"
)

// NewRouter : returns a new Router struct.
func NewRouter(e *gin.Engine, version string, productionEnvironment bool, jwtTokenSecret, latestTermsVersion string, IsAuthTokenValid context.IsAuthTokenValidFunc, getUserAndLocaleFromAuthToken context.GetUserAndLocaleFromAuthTokenFunc, getUserAbilities rights.GetUserAbilitesFunc, getLatestVersionOfTermsAcceptedByUser context.GetLatestVersionOfTermsAcceptedByUserFunc) *Router {
	r := &Router{}

	r.getUserAbilities = getUserAbilities

	r.Group = e.Group("/" + version)
	r.ProductionEnvironment = productionEnvironment

	r.Handlers.CanBeAuthenticated = context.ParseAuthToken(jwtTokenSecret, getUserAndLocaleFromAuthToken, false)
	r.Handlers.MustBeAuthenticated = context.ParseAuthToken(jwtTokenSecret, getUserAndLocaleFromAuthToken, true)
	r.Handlers.MustAuthenticationBeValid = context.VerifyAuthToken(IsAuthTokenValid)
	r.Handlers.MustBeAdmin = r.Rights(true, []string{})
	r.Handlers.AbleToManageUsers = r.Rights(false, []string{rights.AbilityToManageUsers})
	r.Handlers.AbleToManageContacts = r.Rights(false, []string{rights.AbilityToManageContacts})
	r.Handlers.MustHaveAcceptedTheLatestVersionOfTerms = context.VerifyLatestVersionOfTermsAcceptedByUser(latestTermsVersion, getLatestVersionOfTermsAcceptedByUser)

	r.DefaultRoutes.Status.Enabled = true
	r.DefaultRoutes.Status.Paths.Get.Enabled = true
	r.DefaultRoutes.Status.Paths.Get.Path = "/status"

	r.DefaultRoutes.Locales.Enabled = true
	r.DefaultRoutes.Locales.Paths.Get.Enabled = true
	r.DefaultRoutes.Locales.Paths.Get.Path = "/locales"

	r.DefaultRoutes.Timezones.Enabled = true
	r.DefaultRoutes.Timezones.Paths.Get.Enabled = true
	r.DefaultRoutes.Timezones.Paths.Get.Path = "/timezones"

	r.DefaultRoutes.Notifications.Enabled = true
	r.DefaultRoutes.Notifications.Paths.Get.Enabled = true
	r.DefaultRoutes.Notifications.Paths.Get.Path = "/notifications"

	r.DefaultRoutes.Username.Enabled = true
	r.DefaultRoutes.Username.Paths.Availability.Enabled = true
	r.DefaultRoutes.Username.Paths.Availability.Path = "/username/availability"

	r.DefaultRoutes.Terms.Enabled = true
	r.DefaultRoutes.Terms.Paths.GetTerms.Enabled = true
	r.DefaultRoutes.Terms.Paths.GetTerms.Path = "/terms"
	r.DefaultRoutes.Terms.Paths.GetVersion.Enabled = true
	r.DefaultRoutes.Terms.Paths.GetVersion.Path = "/terms/version"
	r.DefaultRoutes.Terms.Paths.Accept.Enabled = true
	r.DefaultRoutes.Terms.Paths.Accept.Path = "/terms/accept"

	r.DefaultRoutes.Registration.Enabled = true
	r.DefaultRoutes.Registration.Paths.RegisterWithPassword.Enabled = true
	r.DefaultRoutes.Registration.Paths.RegisterWithPassword.Path = "/registration"
	r.DefaultRoutes.Registration.Paths.RegisterWithFacebook.Enabled = true
	r.DefaultRoutes.Registration.Paths.RegisterWithFacebook.Path = "/registration/facebook"
	r.DefaultRoutes.Registration.Paths.RegisterWithGoogle.Enabled = true
	r.DefaultRoutes.Registration.Paths.RegisterWithGoogle.Path = "/registration/google"
	r.DefaultRoutes.Registration.Paths.Validate.Enabled = true
	r.DefaultRoutes.Registration.Paths.Validate.Path = "/registration/:id/:key"
	r.DefaultRoutes.Registration.Paths.Cancel.Enabled = true
	r.DefaultRoutes.Registration.Paths.Cancel.Path = "/registration/:id/:key"

	r.DefaultRoutes.ForgotPassword.Enabled = true
	r.DefaultRoutes.ForgotPassword.Paths.ForgotPassword.Enabled = true
	r.DefaultRoutes.ForgotPassword.Paths.ForgotPassword.Path = "/forgotPassword"
	r.DefaultRoutes.ForgotPassword.Paths.Validate.Enabled = true
	r.DefaultRoutes.ForgotPassword.Paths.Validate.Path = "/forgotPassword/:id/:key"
	r.DefaultRoutes.ForgotPassword.Paths.Cancel.Enabled = true
	r.DefaultRoutes.ForgotPassword.Paths.Cancel.Path = "/forgotPassword/:id/:key"

	r.DefaultRoutes.Authentication.Enabled = true
	r.DefaultRoutes.Authentication.Paths.LoginWithPassword.Enabled = true
	r.DefaultRoutes.Authentication.Paths.LoginWithPassword.Path = "/auth/login"
	r.DefaultRoutes.Authentication.Paths.LoginWithFacebook.Enabled = true
	r.DefaultRoutes.Authentication.Paths.LoginWithFacebook.Path = "/auth/login/facebook"
	r.DefaultRoutes.Authentication.Paths.LoginWithGoogle.Enabled = true
	r.DefaultRoutes.Authentication.Paths.LoginWithGoogle.Path = "/auth/login/google"
	r.DefaultRoutes.Authentication.Paths.LoginWithGoogle.Enabled = true
	r.DefaultRoutes.Authentication.Paths.Logout.Enabled = true
	r.DefaultRoutes.Authentication.Paths.Logout.Path = "/auth/logout"
	r.DefaultRoutes.Authentication.Paths.GetTokens.Enabled = true
	r.DefaultRoutes.Authentication.Paths.GetTokens.Path = "/auth/tokens"
	r.DefaultRoutes.Authentication.Paths.GetToken.Enabled = true
	r.DefaultRoutes.Authentication.Paths.GetToken.Path = "/auth/token/:id"
	r.DefaultRoutes.Authentication.Paths.UpdateToken.Enabled = true
	r.DefaultRoutes.Authentication.Paths.UpdateToken.Path = "/auth/token/:id"
	r.DefaultRoutes.Authentication.Paths.RefreshToken.Enabled = true
	r.DefaultRoutes.Authentication.Paths.RefreshToken.Path = "/auth/refreshToken"
	r.DefaultRoutes.Authentication.Paths.DeleteToken.Enabled = true
	r.DefaultRoutes.Authentication.Paths.DeleteToken.Path = "/auth/token/:id"
	r.DefaultRoutes.Authentication.Paths.DeleteTokenWithKey.Enabled = true
	r.DefaultRoutes.Authentication.Paths.DeleteTokenWithKey.Path = "/auth/token/:id/:key"

	r.DefaultRoutes.Account.Enabled = true
	r.DefaultRoutes.Account.Paths.GetNotifications.Enabled = true
	r.DefaultRoutes.Account.Paths.GetNotifications.Path = "/account/notifications"
	r.DefaultRoutes.Account.Paths.GetNotification.Enabled = true
	r.DefaultRoutes.Account.Paths.GetNotification.Path = "/account/notifications/:id"
	r.DefaultRoutes.Account.Paths.DeleteNotification.Enabled = true
	r.DefaultRoutes.Account.Paths.DeleteNotification.Path = "/account/notifications/:id"
	r.DefaultRoutes.Account.Paths.UpdateNotificationRead.Enabled = true
	r.DefaultRoutes.Account.Paths.UpdateNotificationRead.Path = "/account/notifications/:id/read"
	r.DefaultRoutes.Account.Paths.UpdateNotificationNotified.Enabled = true
	r.DefaultRoutes.Account.Paths.UpdateNotificationNotified.Path = "/account/notifications/:id/notified"
	r.DefaultRoutes.Account.Paths.GetNotificationsSettings.Enabled = true
	r.DefaultRoutes.Account.Paths.GetNotificationsSettings.Path = "/account/settings/notifications"
	r.DefaultRoutes.Account.Paths.SaveNotificationsSettings.Enabled = true
	r.DefaultRoutes.Account.Paths.SaveNotificationsSettings.Path = "/account/settings/notifications"
	r.DefaultRoutes.Account.Paths.UnsubscribeFromNotification.Enabled = true
	r.DefaultRoutes.Account.Paths.UnsubscribeFromNotification.Path = "/notifications/unsubscribe/:key"
	r.DefaultRoutes.Account.Paths.GetDetails.Enabled = true
	r.DefaultRoutes.Account.Paths.GetDetails.Path = "/account"
	r.DefaultRoutes.Account.Paths.UpdateDetails.Enabled = true
	r.DefaultRoutes.Account.Paths.UpdateDetails.Path = "/account/infos"
	r.DefaultRoutes.Account.Paths.UpdateSettings.Enabled = true
	r.DefaultRoutes.Account.Paths.UpdateSettings.Path = "/account/settings"
	r.DefaultRoutes.Account.Paths.UpdateSocialNetworks.Enabled = true
	r.DefaultRoutes.Account.Paths.UpdateSocialNetworks.Path = "account/social"
	r.DefaultRoutes.Account.Paths.UpdateEmailAddress.Enabled = true
	r.DefaultRoutes.Account.Paths.UpdateEmailAddress.Path = "/account/emailAddress"
	r.DefaultRoutes.Account.Paths.ValidateUpdateEmailAddress.Enabled = true
	r.DefaultRoutes.Account.Paths.ValidateUpdateEmailAddress.Path = "/account/emailAddress/:id/:key"
	r.DefaultRoutes.Account.Paths.CancelUpdateEmailAddress.Enabled = true
	r.DefaultRoutes.Account.Paths.CancelUpdateEmailAddress.Path = "/account/emailAddress/:id/:key"
	r.DefaultRoutes.Account.Paths.ValidateNewEmailAddress.Enabled = true
	r.DefaultRoutes.Account.Paths.ValidateNewEmailAddress.Path = "/account/validateEmailAddress/:id/:key"
	r.DefaultRoutes.Account.Paths.ConfirmEmailAddress.Enabled = true
	r.DefaultRoutes.Account.Paths.ConfirmEmailAddress.Path = "/account/confirmEmailAddress"
	r.DefaultRoutes.Account.Paths.UpdatePassword.Enabled = true
	r.DefaultRoutes.Account.Paths.UpdatePassword.Path = "/account/password"
	r.DefaultRoutes.Account.Paths.ValidateUpdatePassword.Enabled = true
	r.DefaultRoutes.Account.Paths.ValidateUpdatePassword.Path = "/account/password/:id/:key"
	r.DefaultRoutes.Account.Paths.CancelUpdatePassword.Enabled = true
	r.DefaultRoutes.Account.Paths.CancelUpdatePassword.Path = "/account/password/:id/:key"
	r.DefaultRoutes.Account.Paths.Delete.Enabled = true
	r.DefaultRoutes.Account.Paths.Delete.Path = "/account/delete"
	r.DefaultRoutes.Account.Paths.ValidateDelete.Enabled = true
	r.DefaultRoutes.Account.Paths.ValidateDelete.Path = "/account/delete/:id/:key"
	r.DefaultRoutes.Account.Paths.CancelDelete.Enabled = true
	r.DefaultRoutes.Account.Paths.CancelDelete.Path = "/account/delete/:id/:key"
	r.DefaultRoutes.Account.Paths.Disable.Enabled = true
	r.DefaultRoutes.Account.Paths.Disable.Path = "/account/disable"
	r.DefaultRoutes.Account.Paths.ValidateDisable.Enabled = true
	r.DefaultRoutes.Account.Paths.ValidateDisable.Path = "/account/disable/:id/:key"
	r.DefaultRoutes.Account.Paths.CancelDisable.Enabled = true
	r.DefaultRoutes.Account.Paths.CancelDisable.Path = "/account/disable/:id/:key"

	r.DefaultRoutes.Contact.Enabled = true
	r.DefaultRoutes.Contact.Paths.Post.Enabled = true
	r.DefaultRoutes.Contact.Paths.Post.Path = "/contact"

	r.DefaultRoutes.Admin.Enabled = true
	r.DefaultRoutes.Admin.Paths.GetEnabledUsers.Enabled = true
	r.DefaultRoutes.Admin.Paths.GetEnabledUsers.Path = "/users/enabled"
	r.DefaultRoutes.Admin.Paths.GetDisabledUsers.Enabled = true
	r.DefaultRoutes.Admin.Paths.GetDisabledUsers.Path = "/users/disabled"
	r.DefaultRoutes.Admin.Paths.UpdateUserInformations.Enabled = true
	r.DefaultRoutes.Admin.Paths.UpdateUserInformations.Path = "/users/enabled/:id/informations"
	r.DefaultRoutes.Admin.Paths.UpdateUserDescription.Enabled = true
	r.DefaultRoutes.Admin.Paths.UpdateUserDescription.Path = "/users/enabled/:id/description"
	r.DefaultRoutes.Admin.Paths.UpdateUserSocial.Enabled = true
	r.DefaultRoutes.Admin.Paths.UpdateUserSocial.Path = "/users/enabled/:id/social"
	r.DefaultRoutes.Admin.Paths.UpdateUserSettings.Enabled = true
	r.DefaultRoutes.Admin.Paths.UpdateUserSettings.Path = "/users/enabled/:id/settings"
	r.DefaultRoutes.Admin.Paths.UpdateUserNotifications.Enabled = true
	r.DefaultRoutes.Admin.Paths.UpdateUserNotifications.Path = "/users/enabled/:id/notifications"
	r.DefaultRoutes.Admin.Paths.UpdateUserPassword.Enabled = true
	r.DefaultRoutes.Admin.Paths.UpdateUserPassword.Path = "/users/enabled/:id/password"
	r.DefaultRoutes.Admin.Paths.GetUserTokens.Enabled = true
	r.DefaultRoutes.Admin.Paths.GetUserTokens.Path = "/users/enabled/:id/tokens"
	r.DefaultRoutes.Admin.Paths.DeleteUser.Enabled = true
	r.DefaultRoutes.Admin.Paths.DeleteUser.Path = "/users/enabled/:id"
	r.DefaultRoutes.Admin.Paths.DisableUser.Enabled = true
	r.DefaultRoutes.Admin.Paths.DisableUser.Path = "/users/enabled/:id/disable"
	r.DefaultRoutes.Admin.Paths.EnableUser.Enabled = true
	r.DefaultRoutes.Admin.Paths.EnableUser.Path = "/users/disabled/:id/enable"
	r.DefaultRoutes.Admin.Paths.DeleteUserToken.Enabled = true
	r.DefaultRoutes.Admin.Paths.DeleteUserToken.Path = "/tokens/:id"
	r.DefaultRoutes.Admin.Paths.SetAbilitiesToUser.Enabled = true
	r.DefaultRoutes.Admin.Paths.SetAbilitiesToUser.Path = "/users/enabled/:id/abilities"
	r.DefaultRoutes.Admin.Paths.AddAbilitiesToUser.Enabled = true
	r.DefaultRoutes.Admin.Paths.AddAbilitiesToUser.Path = "/users/enabled/:id/abilities/add"
	r.DefaultRoutes.Admin.Paths.RemoveAbilitiesFromUser.Enabled = true
	r.DefaultRoutes.Admin.Paths.RemoveAbilitiesFromUser.Path = "/users/enabled/:id/abilities/remove"
	r.DefaultRoutes.Admin.Paths.GetContacts.Enabled = true
	r.DefaultRoutes.Admin.Paths.GetContacts.Path = "/contact"
	r.DefaultRoutes.Admin.Paths.GetContact.Enabled = true
	r.DefaultRoutes.Admin.Paths.GetContact.Path = "/contact/:id"
	r.DefaultRoutes.Admin.Paths.UpdateContact.Enabled = true
	r.DefaultRoutes.Admin.Paths.UpdateContact.Path = "/contact/:id"

	return r
}

// Rights : returns the rights for a route.
func (r *Router) Rights(admin bool, abilities []string) gin.HandlerFunc {
	return rights.CheckRights(rights.Rights{AdminRequired: admin, Abilities: abilities}, r.getUserAbilities)
}

// LoadDefaultRoutes : load the default routes.
func (r *Router) LoadDefaultRoutes() {
	// Get the middlewares.
	mustBeAdmin := r.Handlers.MustBeAdmin
	mustBeAuthenticated := r.Handlers.MustBeAuthenticated
	mustAuthenticationBeValid := r.Handlers.MustAuthenticationBeValid
	mustHaveAcceptedTheLatestVersionOfTerms := r.Handlers.MustHaveAcceptedTheLatestVersionOfTerms
	ableToManageUsers := r.Handlers.AbleToManageUsers
	ableToManageContacts := r.Handlers.AbleToManageContacts

	verifyLimitAndOffset := middleware.VerifyLimitAndOffset(0, 20, 100)

	mustBeValidAuthToken := func(messageKey string) gin.HandlerFunc {
		return middleware.VerifyURLParameter(middleware.Config{
			Name:                       "id",
			StoreKey:                   "tokenID",
			MessageKey:                 messageKey,
			NotFoundReasonKey:          "error.admin.token.id.not_found.reason",
			NotFoundRecoveryKey:        "error.admin.token.id.not_found.recovery",
			InvalidObjectIDReasonKey:   "error.admin.token.id.invalid.reason",
			InvalidObjectIDRecoveryKey: "error.admin.token.id.invalid.recovery",
		})
	}

	mustBeValidUserID := func(messageKey string) gin.HandlerFunc {
		return middleware.VerifyURLParameter(middleware.Config{
			Name:                       "id",
			StoreKey:                   "userID",
			MessageKey:                 messageKey,
			NotFoundReasonKey:          "error.admin.user.id.not_found.reason",
			NotFoundRecoveryKey:        "error.admin.user.id.not_found.recovery",
			InvalidObjectIDReasonKey:   "error.admin.user.id.invalid.reason",
			InvalidObjectIDRecoveryKey: "error.admin.user.id.invalid.recovery",
		})
	}

	mustBeValidContactID := func(messageKey string) gin.HandlerFunc {
		return middleware.VerifyURLParameter(middleware.Config{
			Name:                       "id",
			StoreKey:                   "contactID",
			MessageKey:                 messageKey,
			NotFoundReasonKey:          "error.admin.contact.id.not_found.reason",
			NotFoundRecoveryKey:        "error.admin.contact.id.not_found.recovery",
			InvalidObjectIDReasonKey:   "error.admin.contact.id.invalid.reason",
			InvalidObjectIDRecoveryKey: "error.admin.contact.id.invalid.recovery",
		})
	}

	// Set up the default routes.
	v := r.Group

	// ----- Status.
	if r.DefaultRoutes.Status.Enabled {
		if r.DefaultRoutes.Status.Paths.Get.Enabled {
			v.GET(r.DefaultRoutes.Status.Paths.Get.Path, status.GetStatus)
		}
	}

	// ----- Locales.
	if r.DefaultRoutes.Locales.Enabled {
		if r.DefaultRoutes.Locales.Paths.Get.Enabled {
			v.GET(r.DefaultRoutes.Locales.Paths.Get.Path, locale.GetLocales)
		}
	}

	// ----- Timezones.
	if r.DefaultRoutes.Timezones.Enabled {
		if r.DefaultRoutes.Timezones.Paths.Get.Enabled {
			v.GET(r.DefaultRoutes.Timezones.Paths.Get.Path, timezone.GetTimezones)
		}
	}

	// ----- Notifications.
	if r.DefaultRoutes.Notifications.Enabled {
		if r.DefaultRoutes.Notifications.Paths.Get.Enabled {
			supportedNotifications := notificationsSettings.SupportedNotifications()
			v.GET(r.DefaultRoutes.Notifications.Paths.Get.Path, notifications.GetNotifications(supportedNotifications))
		}
	}

	// ----- Username.
	if r.DefaultRoutes.Username.Enabled {
		if r.DefaultRoutes.Username.Paths.Availability.Enabled {
			v.POST(r.DefaultRoutes.Username.Paths.Availability.Path, account.GetAvailability)
		}
	}

	// ----- Terms.
	if r.DefaultRoutes.Terms.Enabled {
		if r.DefaultRoutes.Terms.Paths.GetTerms.Enabled {
			v.GET(r.DefaultRoutes.Terms.Paths.GetTerms.Path, terms.GetTerms)
		}
		if r.DefaultRoutes.Terms.Paths.GetVersion.Enabled {
			v.GET(r.DefaultRoutes.Terms.Paths.GetVersion.Path, terms.GetTermsVersion)
		}
		if r.DefaultRoutes.Terms.Paths.Accept.Enabled {
			v.PUT(r.DefaultRoutes.Terms.Paths.Accept.Path, mustBeAuthenticated, mustAuthenticationBeValid, terms.AcceptTerms)
		}
	}

	// ----- Account : registration.
	if r.DefaultRoutes.Registration.Enabled {
		if r.DefaultRoutes.Registration.Paths.RegisterWithPassword.Enabled {
			v.POST(r.DefaultRoutes.Registration.Paths.RegisterWithPassword.Path, account.RegisterWithPassword)
		}
		if r.DefaultRoutes.Registration.Paths.RegisterWithFacebook.Enabled {
			v.POST(r.DefaultRoutes.Registration.Paths.RegisterWithFacebook.Path, account.RegisterWithFacebook)
		}
		if r.DefaultRoutes.Registration.Paths.RegisterWithGoogle.Enabled {
			v.POST(r.DefaultRoutes.Registration.Paths.RegisterWithGoogle.Path, account.RegisterWithGoogle)
		}
		if r.DefaultRoutes.Registration.Paths.Validate.Enabled {
			v.PUT(r.DefaultRoutes.Registration.Paths.Validate.Path, account.ValidateRegistration)
		}
		if r.DefaultRoutes.Registration.Paths.Cancel.Enabled {
			v.DELETE(r.DefaultRoutes.Registration.Paths.Cancel.Path, account.CancelRegistration)
		}
	}

	// ----- Account : forgot password.
	if r.DefaultRoutes.ForgotPassword.Enabled {
		if r.DefaultRoutes.ForgotPassword.Paths.ForgotPassword.Enabled {
			v.POST(r.DefaultRoutes.ForgotPassword.Paths.ForgotPassword.Path, account.ForgotPassword)
		}
		if r.DefaultRoutes.ForgotPassword.Paths.Validate.Enabled {
			v.PUT(r.DefaultRoutes.ForgotPassword.Paths.Validate.Path, account.ValidateForgotPassword)
		}
		if r.DefaultRoutes.ForgotPassword.Paths.Cancel.Enabled {
			v.DELETE(r.DefaultRoutes.ForgotPassword.Paths.Cancel.Path, account.CancelForgotPassword)
		}
	}

	// ----- Account : auth tokens management.
	if r.DefaultRoutes.Authentication.Enabled {
		if r.DefaultRoutes.Authentication.Paths.LoginWithPassword.Enabled {
			v.POST(r.DefaultRoutes.Authentication.Paths.LoginWithPassword.Path, account.LoginWithPassword)
		}
		if r.DefaultRoutes.Authentication.Paths.LoginWithFacebook.Enabled {
			v.POST(r.DefaultRoutes.Authentication.Paths.LoginWithFacebook.Path, account.LoginWithFacebook)
		}
		if r.DefaultRoutes.Authentication.Paths.LoginWithGoogle.Enabled {
			v.POST(r.DefaultRoutes.Authentication.Paths.LoginWithGoogle.Path, account.LoginWithGoogle)
		}
		if r.DefaultRoutes.Authentication.Paths.Logout.Enabled {
			v.DELETE(r.DefaultRoutes.Authentication.Paths.Logout.Path, mustBeAuthenticated, account.Logout)
		}
		if r.DefaultRoutes.Authentication.Paths.GetTokens.Enabled {
			v.GET(r.DefaultRoutes.Authentication.Paths.GetTokens.Path, mustBeAuthenticated, account.GetAuthTokens)
		}
		if r.DefaultRoutes.Authentication.Paths.GetToken.Enabled {
			v.GET(r.DefaultRoutes.Authentication.Paths.GetToken.Path, mustBeAuthenticated, mustBeValidAuthToken("error.auth_token.get.unable.message"), account.GetAuthToken)
		}
		if r.DefaultRoutes.Authentication.Paths.UpdateToken.Enabled {
			v.PUT(r.DefaultRoutes.Authentication.Paths.UpdateToken.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidAuthToken("error.auth_token.update.unable.message"), account.UpdateAuthToken)
		}
		if r.DefaultRoutes.Authentication.Paths.RefreshToken.Enabled {
			v.POST(r.DefaultRoutes.Authentication.Paths.RefreshToken.Path, mustBeAuthenticated, account.RefreshAuthToken)
		}
		if r.DefaultRoutes.Authentication.Paths.DeleteToken.Enabled {
			v.DELETE(r.DefaultRoutes.Authentication.Paths.DeleteToken.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidAuthToken("error.auth_token.delete.unable.message"), account.DeleteAuthToken)
		}
		if r.DefaultRoutes.Authentication.Paths.DeleteTokenWithKey.Enabled {
			v.DELETE(r.DefaultRoutes.Authentication.Paths.DeleteTokenWithKey.Path, mustBeValidAuthToken("error.auth_token.delete.message"), account.DeleteAuthTokenWithKey)
		}
	}

	// ----- Account : management.
	if r.DefaultRoutes.Account.Enabled {

		// ----- Properties.
		if r.DefaultRoutes.Account.Paths.GetDetails.Enabled {
			v.GET(r.DefaultRoutes.Account.Paths.GetDetails.Path, mustBeAuthenticated, mustAuthenticationBeValid, account.GetInfos)
		}
		if r.DefaultRoutes.Account.Paths.UpdateDetails.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.UpdateDetails.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.UpdateInfos)
		}
		if r.DefaultRoutes.Account.Paths.UpdateSettings.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.UpdateSettings.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.UpdateSettings)
		}
		if r.DefaultRoutes.Account.Paths.UpdateSocialNetworks.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.UpdateSocialNetworks.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.UpdateSocialNetworks)
		}

		// ----- Notification center.
		if r.DefaultRoutes.Account.Paths.GetNotifications.Enabled {
			v.GET(r.DefaultRoutes.Account.Paths.GetNotifications.Path, mustBeAuthenticated, account.GetNotifications)
		}
		if r.DefaultRoutes.Account.Paths.GetNotification.Enabled {
			v.GET(r.DefaultRoutes.Account.Paths.GetNotification.Path, mustBeAuthenticated, account.GetNotification)
		}
		if r.DefaultRoutes.Account.Paths.DeleteNotification.Enabled {
			v.DELETE(r.DefaultRoutes.Account.Paths.DeleteNotification.Path, mustBeAuthenticated, mustAuthenticationBeValid, account.DeleteNotification)
		}
		if r.DefaultRoutes.Account.Paths.UpdateNotificationRead.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.UpdateNotificationRead.Path, mustBeAuthenticated, mustAuthenticationBeValid, account.UpdateNotificationRead)
		}
		if r.DefaultRoutes.Account.Paths.UpdateNotificationNotified.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.UpdateNotificationNotified.Path, mustBeAuthenticated, mustAuthenticationBeValid, account.UpdateNotificationNotified)
		}

		// ----- Notifications.
		if r.DefaultRoutes.Account.Paths.GetNotificationsSettings.Enabled {
			v.GET(r.DefaultRoutes.Account.Paths.GetNotificationsSettings.Path, mustBeAuthenticated, account.GetNotificationsSettings)
		}
		if r.DefaultRoutes.Account.Paths.SaveNotificationsSettings.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.SaveNotificationsSettings.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.SaveNotificationsSettings)
		}
		if r.DefaultRoutes.Account.Paths.UnsubscribeFromNotification.Enabled {
			v.DELETE(r.DefaultRoutes.Account.Paths.UnsubscribeFromNotification.Path, account.UnsubscribeNotificationByKey)
		}

		// ----- Email address.
		if r.DefaultRoutes.Account.Paths.UpdateEmailAddress.Enabled {
			v.POST(r.DefaultRoutes.Account.Paths.UpdateEmailAddress.Path, mustBeAuthenticated, mustHaveAcceptedTheLatestVersionOfTerms, account.UpdateEmailAddress)
		}
		if r.DefaultRoutes.Account.Paths.ValidateUpdateEmailAddress.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.ValidateUpdateEmailAddress.Path, account.ValidateUpdateEmailAddress)
		}
		if r.DefaultRoutes.Account.Paths.CancelUpdateEmailAddress.Enabled {
			v.DELETE(r.DefaultRoutes.Account.Paths.CancelUpdateEmailAddress.Path, account.CancelUpdateEmailAddress)
		}
		if r.DefaultRoutes.Account.Paths.ValidateNewEmailAddress.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.ValidateNewEmailAddress.Path, account.ValidateNewEmailAddress)
		}
		if r.DefaultRoutes.Account.Paths.ConfirmEmailAddress.Enabled {
			v.POST(r.DefaultRoutes.Account.Paths.ConfirmEmailAddress.Path, mustBeAuthenticated, account.ConfirmEmailAddress)
		}

		// ----- Password.
		if r.DefaultRoutes.Account.Paths.UpdatePassword.Enabled {
			v.POST(r.DefaultRoutes.Account.Paths.UpdatePassword.Path, mustBeAuthenticated, mustHaveAcceptedTheLatestVersionOfTerms, account.UpdatePassword)
		}
		if r.DefaultRoutes.Account.Paths.ValidateUpdatePassword.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.ValidateUpdatePassword.Path, account.ValidateUpdatePassword)
		}
		if r.DefaultRoutes.Account.Paths.CancelUpdatePassword.Enabled {
			v.DELETE(r.DefaultRoutes.Account.Paths.CancelUpdatePassword.Path, account.CancelUpdatePassword)
		}

		// ----- Delete.
		if r.DefaultRoutes.Account.Paths.Delete.Enabled {
			v.POST(r.DefaultRoutes.Account.Paths.Delete.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.DeleteAccount)
		}
		if r.DefaultRoutes.Account.Paths.ValidateDelete.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.ValidateDelete.Path, account.ValidateDeleteAccount)
		}
		if r.DefaultRoutes.Account.Paths.CancelDelete.Enabled {
			v.DELETE(r.DefaultRoutes.Account.Paths.CancelDelete.Path, account.CancelDeleteAccount)
		}

		// ----- Disable.
		if r.DefaultRoutes.Account.Paths.Disable.Enabled {
			v.POST(r.DefaultRoutes.Account.Paths.Disable.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustHaveAcceptedTheLatestVersionOfTerms, account.DisableAccount)
		}
		if r.DefaultRoutes.Account.Paths.ValidateDisable.Enabled {
			v.PUT(r.DefaultRoutes.Account.Paths.ValidateDisable.Path, account.ValidateDisableAccount)
		}
		if r.DefaultRoutes.Account.Paths.CancelDisable.Enabled {
			v.DELETE(r.DefaultRoutes.Account.Paths.CancelDisable.Path, account.CancelDisableAccount)
		}
	}

	// ----- Contact
	if r.DefaultRoutes.Contact.Enabled {
		if r.DefaultRoutes.Contact.Paths.Post.Enabled {
			v.POST(r.DefaultRoutes.Contact.Paths.Post.Path, contact.Post)
		}
	}

	// ----- Admin.
	if r.DefaultRoutes.Admin.Enabled {

		// ----- Users.
		if r.DefaultRoutes.Admin.Paths.GetEnabledUsers.Enabled {
			v.GET(r.DefaultRoutes.Admin.Paths.GetEnabledUsers.Path, mustBeAuthenticated, ableToManageUsers, user.GetAdminEnabledUsers)
		}
		if r.DefaultRoutes.Admin.Paths.GetDisabledUsers.Enabled {
			v.GET(r.DefaultRoutes.Admin.Paths.GetDisabledUsers.Path, mustBeAuthenticated, ableToManageUsers, user.GetAdminDisabledUsers)
		}
		if r.DefaultRoutes.Admin.Paths.UpdateUserPassword.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.UpdateUserPassword.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.password.update.message"), ableToManageUsers, user.AdminUpdatePassword)
		}
		if r.DefaultRoutes.Admin.Paths.UpdateUserInformations.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.UpdateUserInformations.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.informations.update.message"), ableToManageUsers, user.AdminUpdateInformations)
		}
		if r.DefaultRoutes.Admin.Paths.UpdateUserDescription.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.UpdateUserDescription.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.description.update.message"), ableToManageUsers, user.AdminUpdateDescription)
		}
		if r.DefaultRoutes.Admin.Paths.UpdateUserSocial.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.UpdateUserSocial.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.social.update.message"), ableToManageUsers, user.AdminUpdateSocial)
		}
		if r.DefaultRoutes.Admin.Paths.UpdateUserSettings.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.UpdateUserSettings.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.settings.update.message"), ableToManageUsers, user.AdminUpdateSettings)
		}
		if r.DefaultRoutes.Admin.Paths.UpdateUserNotifications.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.UpdateUserNotifications.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.notifications.update.message"), ableToManageUsers, user.AdminUpdateNotifications)
		}
		if r.DefaultRoutes.Admin.Paths.GetUserTokens.Enabled {
			v.GET(r.DefaultRoutes.Admin.Paths.GetUserTokens.Path, mustBeAuthenticated, mustBeValidUserID("error.admin.user.tokens.get.message"), ableToManageUsers, user.GetAdminUserTokens)
		}
		if r.DefaultRoutes.Admin.Paths.DeleteUserToken.Enabled {
			v.DELETE(r.DefaultRoutes.Admin.Paths.DeleteUserToken.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidAuthToken("error.admin.user.token.delete.message"), ableToManageUsers, user.AdminDeleteUserToken)
		}
		if r.DefaultRoutes.Admin.Paths.SetAbilitiesToUser.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.SetAbilitiesToUser.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.abilities.update.message"), mustBeAdmin, user.SetAbilities)
		}
		if r.DefaultRoutes.Admin.Paths.AddAbilitiesToUser.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.AddAbilitiesToUser.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.abilities.update.message"), mustBeAdmin, user.AddAbilities)
		}
		if r.DefaultRoutes.Admin.Paths.RemoveAbilitiesFromUser.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.RemoveAbilitiesFromUser.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.abilities.update.message"), mustBeAdmin, user.RemoveAbilities)
		}
		if r.DefaultRoutes.Admin.Paths.DeleteUser.Enabled {
			v.DELETE(r.DefaultRoutes.Admin.Paths.DeleteUser.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.delete.message"), ableToManageUsers, user.AdminDelete)
		}
		if r.DefaultRoutes.Admin.Paths.DisableUser.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.DisableUser.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.disable.message"), ableToManageUsers, user.AdminDisable)
		}
		if r.DefaultRoutes.Admin.Paths.EnableUser.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.EnableUser.Path, mustBeAuthenticated, mustAuthenticationBeValid, mustBeValidUserID("error.admin.user.enable.message"), ableToManageUsers, user.AdminEnable)
		}

		// ----- Contact
		if r.DefaultRoutes.Admin.Paths.GetContacts.Enabled {
			v.GET(r.DefaultRoutes.Admin.Paths.GetContacts.Path, mustBeAuthenticated, ableToManageContacts, verifyLimitAndOffset, contact.GetAdminContacts)
		}
		if r.DefaultRoutes.Admin.Paths.GetContact.Enabled {
			v.GET(r.DefaultRoutes.Admin.Paths.GetContact.Path, mustBeAuthenticated, mustBeValidContactID("error.admin.contact.get.message"), ableToManageContacts, contact.GetAdminContact)
		}
		if r.DefaultRoutes.Admin.Paths.UpdateContact.Enabled {
			v.PUT(r.DefaultRoutes.Admin.Paths.UpdateContact.Path, mustBeAuthenticated, mustBeValidContactID("error.admin.contact.update.message"), ableToManageContacts, contact.Update)
		}
	}

	// ----- Helper routes fo tests/debug.
	if !r.ProductionEnvironment {
		v.GET("/debug/expiredToken", debug.GetExpiredToken)
	}
}
