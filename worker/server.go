// Package worker runs the worker: it consumes the messages of Kafka, to run the asynchronous tasks, send the
// emails and the notifications, and it runs the recurring tasks.
package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/template/text"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/worker/mail"
	"github.com/cygy/ginamite/worker/tasks/async/location"
	"github.com/cygy/ginamite/worker/tasks/recurring"
	"github.com/cygy/ginamite/worker/tasks/recurring/account"
	"github.com/cygy/ginamite/worker/tasks/recurring/authtokens"
	"github.com/cygy/ginamite/worker/tasks/recurring/sitemap"
)

// Server is a worker.
type Server struct {
	// Config is the configuration of the application, set by the main server before it starts.
	Config *config.Configuration
	// Queue sends the messages to the other servers, set by the main server before it starts.
	Queue *queue.Queue
	// Localization translates the emails, set by the main server before it starts.
	Localization *localization.Engine
	RoutesFile   string
	Functions    Handlers

	routes    *weburl.Configuration // The routes of the website, loaded when the worker starts.
	mailer    *mail.Mailer          // Sends the emails, created when the worker starts.
	scheduler *recurring.Scheduler  // Runs the recurring tasks.
}

// Handlers are the functions of a worker, which depend on the storage of the application.
// The worker warns about the undefined functions: the features relying on them are disabled.
type Handlers struct {
	// DeleteExpiredTokens deletes the expired auth tokens and returns their count.
	DeleteExpiredTokens func(ctx context.Context) int64

	// SanitizeAccounts deletes, or updates, the data associated to the deleted accounts and returns their count.
	SanitizeAccounts func() uint

	// DeleteUserByID deletes a user and returns the auth tokens revoked, with their expiration date, by ID.
	DeleteUserByID func(ctx context.Context, userID string) map[string]time.Time

	// DisableUserByID disables a user and returns the auth tokens revoked, with their expiration date, by ID.
	DisableUserByID func(ctx context.Context, userID string) map[string]time.Time

	// EnableUserByID enables a user.
	EnableUserByID func(ctx context.Context, userID string)

	// UpdateUserSocialNetworksByID updates the documents referencing the social networks of a user.
	UpdateUserSocialNetworksByID func(ctx context.Context, userID string)

	// RegistrationDone handles a user who registered.
	RegistrationDone func(ctx context.Context, userID string)

	// RegistrationValidated handles a user who validated their registration.
	RegistrationValidated func(ctx context.Context, userID string)

	// GetNotificationTargetsByUserAndType returns the targets of a user for a type of notification.
	GetNotificationTargetsByUserAndType notifications.GetNotificationTargetsByUserAndTypeFunc

	// SaveIPAddressDetails saves the location of an IP address, returned by getIPAddressDetailsFunc, for an auth token.
	SaveIPAddressDetails func(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *location.IPAddressDetails)

	// StartUpTasks runs once when the worker starts, if the configuration asks for it.
	StartUpTasks func(ctx context.Context, environment, version string)

	// SetFuncOfRecurringTasks sets the functions of the custom recurring tasks with the function it receives.
	SetFuncOfRecurringTasks func(registerTaskFunc func(taskName string, task func(string)))

	// HandleTask handles the custom messages of the topic of the tasks and reports whether it did.
	HandleTask func(ctx context.Context, messageType string, payload []byte) bool

	// HandleUserNotification handles the custom messages of the topic of the user notifications and reports whether it did.
	HandleUserNotification func(ctx context.Context, messageType string, payload []byte, getNotificationTargetsByUserAndTypeFunc notifications.GetNotificationTargetsByUserAndTypeFunc) bool

	// HandleMail handles the custom messages of the topic of the mails and reports whether it did.
	HandleMail func(ctx context.Context, messageType string, payload []byte, isEmailAddressValidFunc mail.IsEmailAddressValidFunc) bool

	// HandleNotificationForUsersWithAbility handles the messages of the topic of the group notifications and reports whether it did.
	HandleNotificationForUsersWithAbility func(ctx context.Context, ability, messageType string, payload []byte) bool

	// IsEmailAddressValid reports whether an email can be sent to an email address.
	IsEmailAddressValid mail.IsEmailAddressValidFunc

	// GetLocalizedRouteVariables returns the variables of the URLs of a route, by locale, from offset, up to limit.
	GetLocalizedRouteVariables func(routeKey string, offset, limit int) map[string][]map[string]string

	// GetNeverUsedAccounts returns the IDs of the accounts which never logged in for intervalInDays days.
	GetNeverUsedAccounts func(ctx context.Context, intervalInDays uint) []string

	// GetInactiveAccounts returns the IDs of the accounts inactive for intervalInMonths months.
	GetInactiveAccounts func(ctx context.Context, intervalInMonths uint) []string

	// IsUserEmpty reports whether a user has no data.
	IsUserEmpty func(ctx context.Context, userID string) bool
}

// NewServer returns a new worker.
func NewServer() *Server {
	return &Server{Config: config.NewConfiguration(), Queue: queue.New(nil), Localization: localization.NewEngine(), scheduler: recurring.NewScheduler()}
}

// Start starts the worker: it loads the templates and the routes, initializes the emails, runs the start up tasks
// and starts the recurring tasks. It returns when the worker is interrupted or terminated, so that the caller can stop
// the consumption of the messages and close the connections.
func (s *Server) Start() {
	s.setup()

	log.Info("started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)
	<-signals

	log.Info("shutdown")
}

// setup prepares and starts the worker, except the consumption of the messages.
func (s *Server) setup() {
	s.warnUndefinedFunctions()

	s.routes = weburl.Load(s.RoutesFile, s.Config.Hosts.Web)

	commonHTMLVars, commonTextVars := s.commonMailVars(), s.commonMailVars()
	s.mailer = mail.New(mail.Settings{
		Username: s.Config.SMTP.Username,
		Password: s.Config.SMTP.Password,
		Host:     s.Config.SMTP.Host,
		Port:     s.Config.SMTP.Port,
		Sender: mail.Sender{
			EmailAddress:        s.Config.SenderEmailAddress,
			Name:                s.Config.SenderName,
			NoReplyEmailAddress: s.Config.NoReplyEmailAddress,
		},
		HTMLVars:      commonHTMLVars,
		TextVars:      commonTextVars,
		Localization:  s.Localization,
		Routes:        s.routes,
		HTMLTemplates: html.LoadTemplatesWithDelimiters(s.Config.TemplatesPath+"/html", "[[", "]]", nil),
		TextTemplates: text.LoadTemplates(s.Config.TemplatesPath + "/text"),
	})

	if s.Functions.StartUpTasks != nil && s.Config.RunStartUpTasks {
		log.Info("check up started")
		s.Functions.StartUpTasks(context.Background(), s.Config.Environment, s.Config.Version)
		log.Info("check up done")
	}

	s.loadRecurringTasks()
	s.scheduler.Start()
}

// commonMailVars returns the variables available in all the emails, by name then by locale.
func (s *Server) commonMailVars() map[string]map[string]any {
	vars := map[string]map[string]any{
		"StaticHost":  {},
		"HomeURL":     {},
		"FacebookURL": {},
		"TwitterURL":  {},
	}

	for _, locale := range s.Config.SupportedLocales {
		vars["StaticHost"][locale] = s.Config.Hosts.Static
		vars["HomeURL"][locale] = fmt.Sprintf("%s/%s/", s.Config.Hosts.Web, locale)
		vars["FacebookURL"][locale] = s.Config.SocialNetworks.Facebook
		vars["TwitterURL"][locale] = s.Config.SocialNetworks.Twitter
	}

	return vars
}

// loadRecurringTasks registers the recurring tasks of the configuration and sets the function of each one.
// A built-in task without the functions it relies on has no function: it is not scheduled.
func (s *Server) loadRecurringTasks() {
	s.scheduler.Load(s.Config.RecurringTasks)

	funcs := s.recurringFuncs()
	for _, task := range s.Config.RecurringTasks {
		if f, ok := funcs[task.Name]; ok {
			s.scheduler.SetFunc(task.Name, f)
		}
	}

	if s.Functions.SetFuncOfRecurringTasks != nil {
		s.Functions.SetFuncOfRecurringTasks(s.scheduler.SetFunc)
	}
}

// recurringFuncs returns the functions of the built-in recurring tasks which can run, by name of task.
func (s *Server) recurringFuncs() map[string]func(taskName string) {
	h := s.Functions
	funcs := map[string]func(taskName string){}

	// Delete the expired auth tokens.
	if h.DeleteExpiredTokens != nil {
		funcs[recurring.DeleteExpiredTokens] = func(taskName string) {
			authtokens.DeleteExpired(context.Background(), taskName, h.DeleteExpiredTokens)
		}
	}

	// Delete, or update, the data associated to the deleted accounts.
	if h.SanitizeAccounts != nil {
		funcs[recurring.SanitizeAccounts] = func(taskName string) {
			account.Sanitize(taskName, h.SanitizeAccounts)
		}
	}

	// Delete the never used accounts.
	if s.Config.Account.EmailAddressMustBeConfirmed && h.GetNeverUsedAccounts != nil && h.IsUserEmpty != nil && h.DeleteUserByID != nil {
		funcs[recurring.DeleteNeverUsedAccounts] = func(taskName string) {
			account.DeleteNeverUsed(context.Background(), s.Queue, taskName, s.Config.Account.DeleteNeverUsedAfter, h.GetNeverUsedAccounts, h.IsUserEmpty, h.DeleteUserByID)
		}
	}

	// Delete the inactive accounts.
	if h.GetInactiveAccounts != nil && h.DeleteUserByID != nil {
		funcs[recurring.DeleteInactiveAccounts] = func(taskName string) {
			account.DeleteInactive(context.Background(), s.Queue, taskName, s.Config.Account.DeleteInactiveAfter, h.GetInactiveAccounts, h.DeleteUserByID)
		}
	}

	// Generate the XML sitemap files.
	if h.GetLocalizedRouteVariables != nil {
		funcs[recurring.GenerateSitemaps] = func(taskName string) {
			sitemap.Generate(taskName, s.Config, s.routes, h.GetLocalizedRouteVariables)
		}
	}

	// The task SendRegistrationMails, sending again the registration emails, is not implemented.

	return funcs
}

// warnUndefinedFunctions logs a warning for each undefined function.
func (s *Server) warnUndefinedFunctions() {
	h := s.Functions

	functions := []struct {
		name    string
		defined bool
	}{
		{"DeleteExpiredTokens", h.DeleteExpiredTokens != nil},
		{"DeleteUserByID", h.DeleteUserByID != nil},
		{"DisableUserByID", h.DisableUserByID != nil},
		{"EnableUserByID", h.EnableUserByID != nil},
		{"UpdateUserSocialNetworksByID", h.UpdateUserSocialNetworksByID != nil},
		{"GetNotificationTargetsByUserAndType", h.GetNotificationTargetsByUserAndType != nil},
		{"SaveIPAddressDetails", h.SaveIPAddressDetails != nil},
		{"SetFuncOfRecurringTasks", h.SetFuncOfRecurringTasks != nil},
		{"StartUpTasks", h.StartUpTasks != nil},
		{"HandleTask", h.HandleTask != nil},
		{"HandleUserNotification", h.HandleUserNotification != nil},
		{"HandleMail", h.HandleMail != nil},
		{"HandleNotificationForUsersWithAbility", h.HandleNotificationForUsersWithAbility != nil},
		{"IsEmailAddressValid", h.IsEmailAddressValid != nil},
		{"GetLocalizedRouteVariables", h.GetLocalizedRouteVariables != nil},
	}

	for _, function := range functions {
		if !function.defined {
			log.Warn(fmt.Sprintf("The function '%s' is undefined.", function.name))
		}
	}
}
