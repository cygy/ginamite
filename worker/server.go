// Package worker runs the worker: it consumes the messages of Kafka, to run the asynchronous tasks, send the
// emails and the notifications, and it runs the recurring tasks.
package worker

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/template/text"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/worker/internal/recurring/account"
	"github.com/cygy/ginamite/worker/internal/recurring/authtokens"
	"github.com/cygy/ginamite/worker/internal/recurring/sitemap"
	"github.com/cygy/ginamite/worker/mail"
	"github.com/cygy/ginamite/worker/tasks/recurring"
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
	// Storage is the storage of the application, such as the one of the storage plugin.
	Storage Storage
	// Hooks are the optional functions of the application.
	Hooks Hooks

	routes    *weburl.Configuration // The routes of the website, loaded when the worker starts.
	mailer    *mail.Mailer          // Sends the emails, created when the worker starts.
	scheduler *recurring.Scheduler  // Runs the recurring tasks.
}

// NewServer returns a new worker.
func NewServer() *Server {
	return &Server{Config: config.NewConfiguration(), Queue: queue.New(nil), Localization: localization.NewEngine(), scheduler: recurring.NewScheduler()}
}

// Run runs the worker: it loads the templates and the routes, initializes the emails, runs the start up tasks
// and starts the recurring tasks. It returns once ctx is done and the running recurring tasks returned, so that the
// caller can stop the consumption of the messages and close the connections, or at once with an error if the set up
// fails.
func (s *Server) Run(ctx context.Context) error {
	if err := s.setup(ctx); err != nil {
		return err
	}

	log.Info("started")
	<-ctx.Done()
	log.Info("shutdown")

	s.scheduler.Wait()

	return nil
}

// setup prepares and starts the worker, except the consumption of the messages. It returns an error if the routes or
// the templates can not be loaded.
func (s *Server) setup(ctx context.Context) error {
	s.warnUndefinedStorages()

	routes, err := weburl.Load(s.RoutesFile, s.Config.Hosts.Web)
	if err != nil {
		return err
	}
	s.routes = routes

	htmlTemplates, err := html.LoadTemplatesWithDelimiters(s.Config.TemplatesPath+"/html", "[[", "]]", nil)
	if err != nil {
		return err
	}
	textTemplates, err := text.LoadTemplates(s.Config.TemplatesPath + "/text")
	if err != nil {
		return err
	}

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
		HTMLTemplates: htmlTemplates,
		TextTemplates: textTemplates,
	})

	if s.Hooks.StartUpTasks != nil && s.Config.RunStartUpTasks {
		log.Info("check up started")
		s.Hooks.StartUpTasks(ctx, s.Config.Environment, s.Config.Version)
		log.Info("check up done")
	}

	s.loadRecurringTasks()
	s.scheduler.Start(ctx)

	return nil
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

	if s.Hooks.SetFuncOfRecurringTasks != nil {
		s.Hooks.SetFuncOfRecurringTasks(s.scheduler.SetFunc)
	}
}

// recurringFuncs returns the functions of the built-in recurring tasks which can run, by name of task.
func (s *Server) recurringFuncs() map[string]recurring.Func {
	st := s.Storage
	funcs := map[string]recurring.Func{}

	// Delete the expired auth tokens.
	if st.Tokens != nil {
		funcs[recurring.DeleteExpiredTokens] = func(ctx context.Context, taskName string) {
			authtokens.DeleteExpired(ctx, taskName, st.Tokens.DeleteExpired)
		}
	}

	// Delete, or update, the data associated to the deleted accounts.
	if st.Accounts != nil {
		funcs[recurring.SanitizeAccounts] = func(ctx context.Context, taskName string) {
			account.Sanitize(ctx, taskName, st.Accounts.Sanitize)
		}
	}

	// Delete the never used accounts.
	if s.Config.Account.EmailAddressMustBeConfirmed && st.Accounts != nil && st.Users != nil {
		funcs[recurring.DeleteNeverUsedAccounts] = func(ctx context.Context, taskName string) {
			account.DeleteNeverUsed(ctx, s.Queue, taskName, s.Config.Account.DeleteNeverUsedAfter, st.Accounts.NeverUsed, st.Accounts.IsEmpty, st.Users.Delete)
		}
	}

	// Delete the inactive accounts.
	if st.Accounts != nil && st.Users != nil {
		funcs[recurring.DeleteInactiveAccounts] = func(ctx context.Context, taskName string) {
			account.DeleteInactive(ctx, s.Queue, taskName, s.Config.Account.DeleteInactiveAfter, st.Accounts.Inactive, st.Users.Delete)
		}
	}

	// Generate the XML sitemap files.
	if st.Sitemap != nil {
		funcs[recurring.GenerateSitemaps] = func(ctx context.Context, taskName string) {
			sitemap.Generate(ctx, taskName, s.Config, s.routes, st.Sitemap.LocalizedRouteVariables)
		}
	}

	return funcs
}

// warnUndefinedStorages logs a warning for each undefined storage, since the features relying on it are disabled.
func (s *Server) warnUndefinedStorages() {
	st := s.Storage

	storages := []struct {
		name    string
		defined bool
	}{
		{"Tokens", st.Tokens != nil},
		{"Accounts", st.Accounts != nil},
		{"Users", st.Users != nil},
		{"Notifications", st.Notifications != nil},
		{"Locations", st.Locations != nil},
		{"Emails", st.Emails != nil},
		{"Sitemap", st.Sitemap != nil},
	}

	for _, storage := range storages {
		if !storage.defined {
			log.Warn(fmt.Sprintf("The storage '%s' is undefined.", storage.name))
		}
	}
}
