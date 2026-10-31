package worker

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/template/text"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/worker/internal/workertest"
	"github.com/cygy/ginamite/worker/mail"
	"github.com/cygy/ginamite/worker/tasks/async/location"
	"github.com/cygy/ginamite/worker/tasks/recurring"
)

// allHandlers returns handlers which do nothing, except the ones of the tasks which log their count.
func allHandlers() Handlers {
	return Handlers{
		DeleteExpiredTokens: func(context.Context) int64 { return 3 },
		SanitizeAccounts:    func() uint { return 4 },
		DeleteUserByID:      func(context.Context, string) map[string]time.Time { return nil },
		GetNeverUsedAccounts: func(_ context.Context, days uint) []string {
			return []string{"never-used"}
		},
		IsUserEmpty: func(context.Context, string) bool { return true },
		GetInactiveAccounts: func(_ context.Context, months uint) []string {
			return []string{"inactive-1", "inactive-2"}
		},
		GetLocalizedRouteVariables: func(string, int, int) map[string][]map[string]string { return nil },
	}
}

func TestNewServer(t *testing.T) {
	if s := NewServer(); s == nil || s.RoutesFile != "" {
		t.Errorf("server = %+v", s)
	}
}

func TestServer_warnUndefinedFunctions(t *testing.T) {
	undefined := func(s *Server) []string {
		workertest.Hook.Reset()
		s.warnUndefinedFunctions()

		names := []string{}
		for _, entry := range workertest.Hook.AllEntries() {
			names = append(names, entry.Message)
		}

		return names
	}

	got := undefined(&Server{})
	if len(got) != 15 {
		t.Fatalf("logged %d warnings, want 15: %v", len(got), got)
	}
	if got[0] != "The function 'DeleteExpiredTokens' is undefined." {
		t.Errorf("first warning = %q", got[0])
	}
	for _, name := range []string{"HandleMail", "IsEmailAddressValid", "GetLocalizedRouteVariables"} {
		if !slices.Contains(got, "The function '"+name+"' is undefined.") {
			t.Errorf("no warning for %s", name)
		}
	}

	h := allHandlers()
	h.DisableUserByID = h.DeleteUserByID
	h.EnableUserByID = func(context.Context, string) {}
	h.UpdateUserSocialNetworksByID = h.EnableUserByID
	h.GetNotificationTargetsByUserAndType = func(context.Context, string, string) ([]notifications.Target, error) { return nil, nil }
	h.SaveIPAddressDetails = func(context.Context, string, string, func(string) *location.IPAddressDetails) {}
	h.SetFuncOfRecurringTasks = func(func(string, func(string))) {}
	h.StartUpTasks = func(context.Context, string, string) {}
	h.HandleTask = func(context.Context, string, []byte) bool { return true }
	h.HandleUserNotification = func(context.Context, string, []byte, notifications.GetNotificationTargetsByUserAndTypeFunc) bool {
		return true
	}
	h.HandleMail = func(context.Context, string, []byte, mail.IsEmailAddressValidFunc) bool { return true }
	h.HandleNotificationForUsersWithAbility = func(context.Context, string, string, []byte) bool { return true }
	h.IsEmailAddressValid = func(context.Context, string) (bool, error) { return true, nil }

	if got := undefined(&Server{Functions: h}); len(got) != 0 {
		t.Errorf("warnings = %v, want none", got)
	}
}

func TestCommonMailVars(t *testing.T) {
	c := &config.Configuration{SupportedLocales: []string{"en", "fr"}}
	c.Hosts.Web = "https://web"
	c.Hosts.Static = "https://static"
	c.SocialNetworks.Facebook = "https://facebook"
	c.SocialNetworks.Twitter = "https://twitter"

	vars := (&Server{Config: c}).commonMailVars()

	want := map[string]map[string]any{
		"StaticHost":  {"en": "https://static", "fr": "https://static"},
		"HomeURL":     {"en": "https://web/en/", "fr": "https://web/fr/"},
		"FacebookURL": {"en": "https://facebook", "fr": "https://facebook"},
		"TwitterURL":  {"en": "https://twitter", "fr": "https://twitter"},
	}
	if len(vars) != len(want) {
		t.Fatalf("vars = %v", vars)
	}
	for name, byLocale := range want {
		for locale, value := range byLocale {
			if vars[name][locale] != value {
				t.Errorf("%s[%s] = %v, want %v", name, locale, vars[name][locale], value)
			}
		}
	}

	// Each call returns independent maps.
	vars["HomeURL"]["en"] = "changed"
	if (&Server{Config: c}).commonMailVars()["HomeURL"]["en"] != "https://web/en/" {
		t.Error("the variables are shared between the calls")
	}
}

func TestServer_recurringFuncs(t *testing.T) {
	names := func(s *Server) []string {
		keys := []string{}
		for name := range s.recurringFuncs() {
			keys = append(keys, name)
		}
		sort.Strings(keys)

		return keys
	}

	t.Run("no function", func(t *testing.T) {
		if got := names(&Server{Config: &config.Configuration{}}); len(got) != 0 {
			t.Errorf("tasks = %v, want none", got)
		}
	})

	t.Run("all the functions", func(t *testing.T) {
		c := &config.Configuration{}
		c.Account.EmailAddressMustBeConfirmed = true

		want := []string{recurring.DeleteExpiredTokens, recurring.DeleteInactiveAccounts, recurring.DeleteNeverUsedAccounts, recurring.GenerateSitemaps, recurring.SanitizeAccounts}
		sort.Strings(want)
		if got := names(&Server{Config: c, Functions: allHandlers()}); !slices.Equal(got, want) {
			t.Errorf("tasks = %v, want %v", got, want)
		}
	})

	t.Run("the never used accounts are kept if the email address does not need to be confirmed", func(t *testing.T) {
		if got := names(&Server{Config: &config.Configuration{}, Functions: allHandlers()}); slices.Contains(got, recurring.DeleteNeverUsedAccounts) {
			t.Errorf("tasks = %v", got)
		}
	})

	t.Run("missing dependencies", func(t *testing.T) {
		c := &config.Configuration{}
		c.Account.EmailAddressMustBeConfirmed = true

		h := allHandlers()
		h.DeleteUserByID = nil
		h.GetLocalizedRouteVariables = nil

		if got, want := names(&Server{Config: c, Functions: h}), []string{recurring.DeleteExpiredTokens, recurring.SanitizeAccounts}; !slices.Equal(got, want) {
			t.Errorf("tasks = %v, want %v", got, want)
		}
	})
}

func TestServer_recurringFuncs_run(t *testing.T) {
	c := &config.Configuration{}
	c.Account.EmailAddressMustBeConfirmed = true
	c.Account.DeleteNeverUsedAfter = 10
	c.Account.DeleteInactiveAfter = 48
	c.Sitemap = config.SitemapConfiguration{Path: t.TempDir(), Name: "sitemap", MaximumURLsPerPage: 1000}

	oldRoutes := weburl.Main
	t.Cleanup(func() { weburl.Main = oldRoutes })
	weburl.Main = weburl.NewConfiguration()

	workertest.CaptureQueue(t)
	s := &Server{Config: c, Functions: allHandlers()}

	// The count logged by each task.
	wantCounts := map[string]any{
		recurring.DeleteExpiredTokens:     int64(3),
		recurring.SanitizeAccounts:        uint(4),
		recurring.DeleteNeverUsedAccounts: 1,
		recurring.DeleteInactiveAccounts:  2,
	}

	for name, f := range s.recurringFuncs() {
		t.Run(name, func(t *testing.T) {
			workertest.Hook.Reset()
			f(name)

			var done map[string]any
			for _, entry := range workertest.Hook.AllEntries() {
				if entry.Message == "task done" {
					done = entry.Data
				}
			}

			if done == nil || done["task"] != name {
				t.Fatalf("the task is not logged as done: %v", done)
			}
			if want, ok := wantCounts[name]; ok && done["count"] != want {
				t.Errorf("count = %v, want %v", done["count"], want)
			}
		})
	}
}

func TestServer_loadRecurringTasks(t *testing.T) {
	c := &config.Configuration{}
	c.RecurringTasks = []recurring.ScheduledTask{
		{Name: recurring.DeleteExpiredTokens},
		{Name: recurring.SanitizeAccounts},
		{Name: recurring.SendRegistrationMails},
		{Name: "custom"},
	}

	h := Handlers{DeleteExpiredTokens: allHandlers().DeleteExpiredTokens}
	var custom bool
	h.SetFuncOfRecurringTasks = func(register func(taskName string, task func(string))) {
		register("custom", func(string) {})
		custom = true
	}

	(&Server{Config: c, Functions: h}).loadRecurringTasks()

	if !custom {
		t.Error("the functions of the custom tasks are not set")
	}
	for name, want := range map[string]bool{
		recurring.DeleteExpiredTokens:   true,  // The function is defined.
		recurring.SanitizeAccounts:      false, // The function is not defined.
		recurring.SendRegistrationMails: false, // The task has no function.
		"custom":                        true,  // The function is set by SetFuncOfRecurringTasks.
		"unknown":                       false,
	} {
		if got := recurring.HasFunc(name); got != want {
			t.Errorf("HasFunc(%s) = %v, want %v", name, got, want)
		}
	}

	// Without a function to set the custom tasks.
	(&Server{Config: c}).loadRecurringTasks()
}

func TestServer_setup(t *testing.T) {
	dir := t.TempDir()
	write := func(path, content string) {
		t.Helper()

		path = filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("templates/html/mails/hello.html", "<p>[[.Name]]</p>")
	write("templates/text/mails/hello.txt", "Hello {{.Name}}")
	write("routes.yml", "routes:\n  home:\n    paths:\n      en: /en/\n")

	c := &config.Configuration{Environment: "test", Version: "1.0", TemplatesPath: filepath.Join(dir, "templates"), RunStartUpTasks: true}
	c.Hosts.Web = "https://web"
	c.SupportedLocales = []string{"en"}
	c.RecurringTasks = []recurring.ScheduledTask{{Name: recurring.DeleteExpiredTokens}}

	oldHTML, oldText, oldRoutes := html.Main, text.Main, weburl.Main
	t.Cleanup(func() { html.Main, text.Main, weburl.Main = oldHTML, oldText, oldRoutes })

	var env, version string
	var calls []string
	s := &Server{Config: c, RoutesFile: filepath.Join(dir, "routes.yml")}
	s.Functions.StartUpTasks = func(_ context.Context, environment, v string) {
		env, version = environment, v
		calls = append(calls, "start up")
	}
	s.Functions.SetFuncOfRecurringTasks = func(func(string, func(string))) { calls = append(calls, "recurring") }
	s.Functions.DeleteExpiredTokens = allHandlers().DeleteExpiredTokens

	s.setup()

	if env != "test" || version != "1.0" || !slices.Equal(calls, []string{"start up", "recurring"}) {
		t.Errorf("env = %q, version = %q, calls = %v", env, version, calls)
	}
	if html.Main == nil || text.Main == nil {
		t.Fatal("the templates are not loaded")
	}

	var b strings.Builder
	if err := html.Main.ExecuteTemplate(&b, "hello.html", map[string]string{"Name": "John"}); err != nil || b.String() != "<p>John</p>" {
		t.Errorf("HTML template = %q, %v", b.String(), err)
	}
	if got := weburl.GetURL("home", "en"); got != "https://web/en/" {
		t.Errorf("URL of the home = %q", got)
	}
	if !recurring.HasFunc(recurring.DeleteExpiredTokens) {
		t.Error("the recurring tasks are not loaded")
	}

	// The start up tasks only run if the configuration asks for it.
	c.RunStartUpTasks = false
	calls = nil
	s.setup()
	if slices.Contains(calls, "start up") {
		t.Errorf("calls = %v", calls)
	}
}
