package worker

import (
	"time"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/worker/mail"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

// Server : definition of a worker server.
type Server struct {
	RoutesFile string
	Functions  Handlers
}

// Handlers : list of the functions of a worker
type Handlers struct {
	DeleteExpiredTokens                   func() uint
	DeleteUserByID                        func(userID string) map[string]time.Time
	DisableUserByID                       func(userID string) map[string]time.Time
	EnableUserByID                        func(userID string)
	UpdateUserSocialNetworksByID          func(userID string)
	RegistrationDone                      func(userID string)
	RegistrationValidated                 func(userID string)
	GetNotificationTargetsByUserAndType   notifications.GetNotificationTargetsByUserAndTypeFunc
	SaveIPAddressDetails                  func(IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *location.IPAddressDetails)
	StartUpTasks                          func(environment, version string)
	SetFuncOfRecurringTasks               func(registerTaskFunc func(taskName string, task func(string)))
	HandleTask                            func(messageType string, payload []byte) bool
	HandleUserNotification                func(messageType string, payload []byte, getNotificationTargetsByUserAndTypeFunc notifications.GetNotificationTargetsByUserAndTypeFunc) bool
	HandleMail                            func(messageType string, payload []byte, isEmailAddressValidFunc mail.IsEmailAddressValidFunc) bool
	HandleNotificationForUsersWithAbility func(ability, messageType string, payload []byte) bool
	IsEmailAddressValid                   mail.IsEmailAddressValidFunc
	GetLocalizedRouteVariables            func(routeKey string, offset, limit int) map[string][]map[string]string // indexed by locale, then indexed by variable name.
	GetNeverUsedAccounts                  func(intervalInDays uint) []string
	GetInactiveAccounts                   func(intervalInMonths uint) []string
}
