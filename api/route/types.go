// Package route defines the routes of the API and the middlewares applied to them.
package route

import (
	"github.com/cygy/ginamite/api/rights"
	"github.com/cygy/ginamite/common/notifications"

	"github.com/gin-gonic/gin"
)

// ConfigureDefaultRoutes is a function to configure the default routes of the API.
type ConfigureDefaultRoutes func(routes *DefaultRoutes)

// ConfigureCustomRoutes is a function to add custom routes of the API.
type ConfigureCustomRoutes func(group *gin.RouterGroup, handlers Middlewares, isProductionEnvironment bool, rights func(admin bool, abilities []string) gin.HandlerFunc)

// Router holds the properties to define routes and their rights.
type Router struct {
	Group                 *gin.RouterGroup
	ProductionEnvironment bool
	DefaultRoutes         DefaultRoutes
	Handlers              Middlewares

	// Limits are the middlewares limiting the requests which try to guess a secret. A nil one does not limit.
	Limits RateLimits

	// Notifications is the registry of the notifications the users can subscribe to.
	Notifications *notifications.Registry

	// private properties
	getUserAbilities       rights.GetUserAbilitiesFunc
	authenticationDisabled bool
}

// RateLimits are the middlewares limiting the requests of the default routes, by group of routes.
type RateLimits struct {
	Authentication gin.HandlerFunc
	Registration   gin.HandlerFunc
	ForgotPassword gin.HandlerFunc
}

// Middlewares is the collection of the built-in handlers.
type Middlewares struct {
	CanBeAuthenticated                      gin.HandlerFunc
	MustBeAuthenticated                     gin.HandlerFunc
	MustAuthenticationBeValid               gin.HandlerFunc
	MustBeAdmin                             gin.HandlerFunc
	AbleToManageUsers                       gin.HandlerFunc
	AbleToManageContacts                    gin.HandlerFunc
	MustHaveAcceptedTheLatestVersionOfTerms gin.HandlerFunc
}

// PathProperties holds the properties of a route.
type PathProperties struct {
	Enabled bool
	Path    string
}

// DefaultRoutes defines the built-in routes.
type DefaultRoutes struct {
	Status struct {
		Enabled bool
		Paths   struct {
			Get PathProperties
		}
	}
	Locales struct {
		Enabled bool
		Paths   struct {
			Get PathProperties
		}
	}
	Timezones struct {
		Enabled bool
		Paths   struct {
			Get PathProperties
		}
	}
	Notifications struct {
		Enabled bool
		Paths   struct {
			Get PathProperties
		}
	}
	Username struct {
		Enabled bool
		Paths   struct {
			Availability PathProperties
		}
	}
	Terms struct {
		Enabled bool
		Paths   struct {
			GetTerms   PathProperties
			GetVersion PathProperties
			Accept     PathProperties
		}
	}
	Registration struct {
		Enabled bool
		Paths   struct {
			RegisterWithPassword PathProperties
			RegisterWithFacebook PathProperties
			RegisterWithGoogle   PathProperties
			Validate             PathProperties
			Cancel               PathProperties
		}
	}
	ForgotPassword struct {
		Enabled bool
		Paths   struct {
			ForgotPassword PathProperties
			Validate       PathProperties
			Cancel         PathProperties
		}
	}
	Authentication struct {
		Enabled bool
		Paths   struct {
			LoginWithPassword  PathProperties
			LoginWithFacebook  PathProperties
			LoginWithGoogle    PathProperties
			Logout             PathProperties
			GetTokens          PathProperties
			GetToken           PathProperties
			UpdateToken        PathProperties
			RefreshToken       PathProperties
			DeleteToken        PathProperties
			DeleteTokenWithKey PathProperties
		}
	}
	Account struct {
		Enabled bool
		Paths   struct {
			GetNotifications            PathProperties
			GetNotification             PathProperties
			DeleteNotification          PathProperties
			UpdateNotificationRead      PathProperties
			UpdateNotificationNotified  PathProperties
			GetNotificationsSettings    PathProperties
			SaveNotificationsSettings   PathProperties
			UnsubscribeFromNotification PathProperties
			GetDetails                  PathProperties
			UpdateDetails               PathProperties
			UpdateSettings              PathProperties
			UpdateSocialNetworks        PathProperties
			UpdateEmailAddress          PathProperties
			ValidateUpdateEmailAddress  PathProperties
			CancelUpdateEmailAddress    PathProperties
			ValidateNewEmailAddress     PathProperties
			ConfirmEmailAddress         PathProperties
			UpdatePassword              PathProperties
			ValidateUpdatePassword      PathProperties
			CancelUpdatePassword        PathProperties
			Delete                      PathProperties
			ValidateDelete              PathProperties
			CancelDelete                PathProperties
			Disable                     PathProperties
			ValidateDisable             PathProperties
			CancelDisable               PathProperties
		}
	}
	Contact struct {
		Enabled bool
		Paths   struct {
			Post PathProperties
		}
	}
	Admin struct {
		Enabled bool
		Paths   struct {
			GetEnabledUsers         PathProperties
			GetDisabledUsers        PathProperties
			UpdateUserInformations  PathProperties
			UpdateUserDescription   PathProperties
			UpdateUserSocial        PathProperties
			UpdateUserSettings      PathProperties
			UpdateUserNotifications PathProperties
			UpdateUserPassword      PathProperties
			GetUserTokens           PathProperties
			DeleteUserToken         PathProperties
			DeleteUser              PathProperties
			DisableUser             PathProperties
			EnableUser              PathProperties
			SetAbilitiesToUser      PathProperties
			AddAbilitiesToUser      PathProperties
			RemoveAbilitiesFromUser PathProperties
			GetContacts             PathProperties
			GetContact              PathProperties
			UpdateContact           PathProperties
		}
	}
}
