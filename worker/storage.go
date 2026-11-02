package worker

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/worker/mail"
	"github.com/cygy/ginamite/worker/tasks/async/location"
	"github.com/cygy/ginamite/worker/tasks/recurring"
)

// Storage is the storage of the application used by a worker, by feature: a database plugin implements the interfaces.
// An undefined storage disables the features relying on it, and the worker warns about it when it starts.
//
// The methods handling a message of Kafka return an error if they failed: the message is then handled again, then
// dead-lettered if the error persists, see broker.MessageConsumer. An error marked with broker.Permanent is not
// retried. They must tolerate a message handled twice.
type Storage struct {
	// Tokens is required by the recurring task deleting the expired auth tokens.
	Tokens TokensStorage
	// Accounts is required by the recurring tasks about the accounts.
	Accounts AccountsStorage
	// Users is required by the tasks deleting, disabling and enabling a user.
	Users UsersStorage
	// Notifications is required by the notifications sent to a user.
	Notifications NotificationsStorage
	// Locations is required by the task saving the location of an IP address.
	Locations LocationsStorage
	// Emails is required to send an email.
	Emails EmailsStorage
	// Sitemap is required by the recurring task generating the sitemaps of the routes with variables.
	Sitemap SitemapStorage
}

// TokensStorage is the storage of the auth tokens.
type TokensStorage interface {
	// DeleteExpired deletes the expired auth tokens and returns their count.
	DeleteExpired(ctx context.Context) (int64, error)
}

// AccountsStorage is the storage of the accounts used by the recurring tasks.
type AccountsStorage interface {
	// NeverUsed returns the IDs of the accounts which never logged in for intervalInDays days.
	NeverUsed(ctx context.Context, intervalInDays uint) ([]string, error)
	// Inactive returns the IDs of the accounts inactive for intervalInMonths months.
	Inactive(ctx context.Context, intervalInMonths uint) ([]string, error)
	// IsEmpty reports whether a user has no data.
	IsEmpty(ctx context.Context, userID string) (bool, error)
	// Sanitize deletes, or updates, the data associated to the deleted accounts and returns their count.
	Sanitize(ctx context.Context) (uint, error)
}

// UsersStorage is the storage of the users used by the tasks which follow the changes of an account.
type UsersStorage interface {
	// Delete deletes a user and returns the auth tokens revoked, with their expiration date, by ID.
	Delete(ctx context.Context, userID string) (map[string]time.Time, error)
	// Disable disables a user and returns the auth tokens revoked, with their expiration date, by ID.
	Disable(ctx context.Context, userID string) (map[string]time.Time, error)
	// Enable enables a user.
	Enable(ctx context.Context, userID string) error
	// UpdateSocialNetworks updates the documents referencing the social networks of a user.
	UpdateSocialNetworks(ctx context.Context, userID string) error
}

// NotificationsStorage is the storage of the targets of the notifications.
type NotificationsStorage interface {
	// Targets returns the targets of a user for a type of notification.
	Targets(ctx context.Context, userID, notificationType string) ([]notifications.Target, error)
}

// LocationsStorage is the storage of the locations of the IP addresses.
type LocationsStorage interface {
	// SaveIPAddressDetails saves the location of an IP address, returned by getDetails, for an auth token.
	SaveIPAddressDetails(ctx context.Context, ipAddress, tokenID string, getDetails location.DetailsFunc) error
}

// EmailsStorage is the storage of the email addresses.
type EmailsStorage interface {
	// IsValid reports whether an email can be sent to an email address.
	IsValid(ctx context.Context, emailAddress string) (bool, error)
}

// SitemapStorage is the storage of the data of the sitemaps.
type SitemapStorage interface {
	// LocalizedRouteVariables returns the variables of the URLs of a route, by locale, from offset, up to limit.
	LocalizedRouteVariables(ctx context.Context, routeKey string, offset, limit int) (map[string][]map[string]string, error)
}

// Hooks are the functions of the application which are not tied to a storage: the start up tasks, the custom tasks and
// the handling of the custom messages. They are all optional.
type Hooks struct {
	// StartUpTasks runs once when the worker starts, if the configuration asks for it.
	StartUpTasks func(ctx context.Context, environment, version string)

	// SetFuncOfRecurringTasks sets the functions of the custom recurring tasks with the function it receives. A task
	// receives the context of the worker, done when it stops.
	SetFuncOfRecurringTasks func(registerTaskFunc func(taskName string, task recurring.Func))

	// RegistrationDone handles a user who registered.
	RegistrationDone func(ctx context.Context, userID string) error

	// RegistrationValidated handles a user who validated their registration.
	RegistrationValidated func(ctx context.Context, userID string) error

	// HandleTask handles the custom messages of the topic of the tasks and reports whether it did, and its error.
	HandleTask func(ctx context.Context, messageType string, payload []byte) (bool, error)

	// HandleUserNotification handles the custom messages of the topic of the user notifications and reports whether it
	// did, and its error. It receives the function returning the targets of a user, nil if the storage is undefined.
	HandleUserNotification func(ctx context.Context, messageType string, payload []byte, getTargets notifications.GetNotificationTargetsByUserAndTypeFunc) (bool, error)

	// HandleMail handles the custom messages of the topic of the mails and reports whether it did, and its error. It
	// receives the function checking an email address, nil if the storage is undefined.
	HandleMail func(ctx context.Context, messageType string, payload []byte, isEmailAddressValid mail.IsEmailAddressValidFunc) (bool, error)

	// HandleNotificationForUsersWithAbility handles the messages of the topic of the group notifications and reports
	// whether it did, and its error.
	HandleNotificationForUsersWithAbility func(ctx context.Context, ability, messageType string, payload []byte) (bool, error)
}

// targetsFunc returns the function reading the targets of the notifications, nil if the storage is undefined.
func (s *Storage) targetsFunc() notifications.GetNotificationTargetsByUserAndTypeFunc {
	if s.Notifications == nil {
		return nil
	}

	return s.Notifications.Targets
}

// emailCheckFunc returns the function checking the email addresses, nil if the storage is undefined.
func (s *Storage) emailCheckFunc() mail.IsEmailAddressValidFunc {
	if s.Emails == nil {
		return nil
	}

	return s.Emails.IsValid
}

// deleteUserFunc returns the function deleting a user, nil if the storage is undefined.
func (s *Storage) deleteUserFunc() func(ctx context.Context, userID string) (map[string]time.Time, error) {
	if s.Users == nil {
		return nil
	}

	return s.Users.Delete
}

// disableUserFunc returns the function disabling a user, nil if the storage is undefined.
func (s *Storage) disableUserFunc() func(ctx context.Context, userID string) (map[string]time.Time, error) {
	if s.Users == nil {
		return nil
	}

	return s.Users.Disable
}

// enableUserFunc returns the function enabling a user, nil if the storage is undefined.
func (s *Storage) enableUserFunc() func(ctx context.Context, userID string) error {
	if s.Users == nil {
		return nil
	}

	return s.Users.Enable
}

// updateSocialNetworksFunc returns the function updating the social networks of a user, nil if the storage is undefined.
func (s *Storage) updateSocialNetworksFunc() func(ctx context.Context, userID string) error {
	if s.Users == nil {
		return nil
	}

	return s.Users.UpdateSocialNetworks
}

// saveIPAddressDetailsFunc returns the function saving the location of an IP address, nil if the storage is undefined.
func (s *Storage) saveIPAddressDetailsFunc() func(ctx context.Context, ipAddress, tokenID string, getDetails location.DetailsFunc) error {
	if s.Locations == nil {
		return nil
	}

	return s.Locations.SaveIPAddressDetails
}
