package route

import (
	"github.com/cygy/ginamite/api/rights"

	"github.com/gin-gonic/gin"
)

// ConfigureDefaultRoutes : function to configure the default routes of the API.
type ConfigureDefaultRoutes func(routes *DefaultRoutes)

// ConfigureCustomRoutes : function to add custom routes of the API.
type ConfigureCustomRoutes func(group *gin.RouterGroup, handlers Middlewares, isProductionEnvironment bool, rights func(admin bool, abilities []string) gin.HandlerFunc)

// Router : properties to define routes and their rights.
type Router struct {
	Group                 *gin.RouterGroup
	ProductionEnvironment bool
	DefaultRoutes         DefaultRoutes
	Handlers              Middlewares

	// private properties
	getUserAbilities rights.GetUserAbilitesFunc
}

// Middlewares : collection of the built-in handlers.
type Middlewares struct {
	CanBeAuthenticated                      gin.HandlerFunc
	MustBeAuthenticated                     gin.HandlerFunc
	MustAuthenticationBeValid               gin.HandlerFunc
	MustBeAdmin                             gin.HandlerFunc
	AbleToManageUsers                       gin.HandlerFunc
	AbleToManageContacts                    gin.HandlerFunc
	MustHaveAcceptedTheLatestVersionOfTerms gin.HandlerFunc
}

// PathProperties : properties of a route.
type PathProperties struct {
	Enabled bool
	Path    string
}

// DefaultRoutes : definition of the built-in routes.
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
