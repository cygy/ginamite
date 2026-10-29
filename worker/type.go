package worker

import (
	"context"
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
	DeleteExpiredTokens                   func(ctx context.Context) int64
	SanitizeAccounts                      func() uint
	DeleteUserByID                        func(ctx context.Context, userID string) map[string]time.Time
	DisableUserByID                       func(ctx context.Context, userID string) map[string]time.Time
	EnableUserByID                        func(ctx context.Context, userID string)
	UpdateUserSocialNetworksByID          func(ctx context.Context, userID string)
	RegistrationDone                      func(ctx context.Context, userID string)
	RegistrationValidated                 func(ctx context.Context, userID string)
	GetNotificationTargetsByUserAndType   notifications.GetNotificationTargetsByUserAndTypeFunc
	SaveIPAddressDetails                  func(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *location.IPAddressDetails)
	StartUpTasks                          func(ctx context.Context, environment, version string)
	SetFuncOfRecurringTasks               func(registerTaskFunc func(taskName string, task func(string)))
	HandleTask                            func(ctx context.Context, messageType string, payload []byte) bool
	HandleUserNotification                func(ctx context.Context, messageType string, payload []byte, getNotificationTargetsByUserAndTypeFunc notifications.GetNotificationTargetsByUserAndTypeFunc) bool
	HandleMail                            func(ctx context.Context, messageType string, payload []byte, isEmailAddressValidFunc mail.IsEmailAddressValidFunc) bool
	HandleNotificationForUsersWithAbility func(ctx context.Context, ability, messageType string, payload []byte) bool
	IsEmailAddressValid                   mail.IsEmailAddressValidFunc
	GetLocalizedRouteVariables            func(routeKey string, offset, limit int) map[string][]map[string]string // indexed by locale, then indexed by variable name.
	GetNeverUsedAccounts                  func(ctx context.Context, intervalInDays uint) []string
	GetInactiveAccounts                   func(ctx context.Context, intervalInMonths uint) []string
	IsUserEmpty                           func(ctx context.Context, userID string) bool
}
