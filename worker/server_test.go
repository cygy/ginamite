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
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/worker/internal/workertest"
	"github.com/cygy/ginamite/worker/tasks/recurring"
)

func TestNewServer(t *testing.T) {
	if s := NewServer(); s == nil || s.Routes != nil {
		t.Errorf("server = %+v", s)
	}
}

func TestServer_warnUndefinedStorages(t *testing.T) {
	undefined := func(s *Server) []string {
		workertest.Hook.Reset()
		s.warnUndefinedStorages()

		names := []string{}
		for _, entry := range workertest.Hook.AllEntries() {
			names = append(names, entry.Message)
		}

		return names
	}

	got := undefined(&Server{})
	if len(got) != 7 {
		t.Fatalf("logged %d warnings, want 7: %v", len(got), got)
	}
	if got[0] != "The storage 'Tokens' is undefined." {
		t.Errorf("first warning = %q", got[0])
	}
	for _, name := range []string{"Emails", "Sitemap", "Locations"} {
		if !slices.Contains(got, "The storage '"+name+"' is undefined.") {
			t.Errorf("no warning for %s", name)
		}
	}

	if got := undefined(&Server{Storage: allStorages()}); len(got) != 0 {
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

	t.Run("no storage", func(t *testing.T) {
		if got := names(&Server{Config: &config.Configuration{}}); len(got) != 0 {
			t.Errorf("tasks = %v, want none", got)
		}
	})

	t.Run("all the storages", func(t *testing.T) {
		c := &config.Configuration{}
		c.Account.EmailAddressMustBeConfirmed = true

		want := []string{recurring.DeleteExpiredTokens, recurring.DeleteInactiveAccounts, recurring.DeleteNeverUsedAccounts, recurring.GenerateSitemaps, recurring.SanitizeAccounts}
		sort.Strings(want)
		if got := names(&Server{Config: c, Routes: weburl.NewConfiguration(), Storage: allStorages()}); !slices.Equal(got, want) {
			t.Errorf("tasks = %v, want %v", got, want)
		}
	})

	t.Run("the sitemaps need the web routes", func(t *testing.T) {
		if got := names(&Server{Config: &config.Configuration{}, Storage: allStorages()}); slices.Contains(got, recurring.GenerateSitemaps) {
			t.Errorf("tasks = %v", got)
		}
	})

	t.Run("the never used accounts are kept if the email address does not need to be confirmed", func(t *testing.T) {
		if got := names(&Server{Config: &config.Configuration{}, Storage: allStorages()}); slices.Contains(got, recurring.DeleteNeverUsedAccounts) {
			t.Errorf("tasks = %v", got)
		}
	})

	t.Run("missing storages", func(t *testing.T) {
		c := &config.Configuration{}
		c.Account.EmailAddressMustBeConfirmed = true

		st := allStorages()
		st.Users = nil
		st.Sitemap = nil

		if got, want := names(&Server{Config: c, Storage: st}), []string{recurring.DeleteExpiredTokens, recurring.SanitizeAccounts}; !slices.Equal(got, want) {
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

	q := workertest.CaptureQueue(t)
	s := &Server{Config: c, Queue: q.Sender, Routes: weburl.NewConfiguration(), Storage: allStorages()}

	// The count logged by each task.
	wantCounts := map[string]any{
		recurring.DeleteExpiredTokens:     3,
		recurring.SanitizeAccounts:        uint(4),
		recurring.DeleteNeverUsedAccounts: 1,
		recurring.DeleteInactiveAccounts:  2,
	}

	for name, f := range s.recurringFuncs() {
		t.Run(name, func(t *testing.T) {
			workertest.Hook.Reset()
			f(t.Context(), name)

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
	c.RecurringTasks = []config.RecurringTask{
		{Name: recurring.DeleteExpiredTokens},
		{Name: recurring.SanitizeAccounts},
		{Name: "custom"},
	}

	h := Hooks{}
	var custom bool
	h.SetFuncOfRecurringTasks = func(register func(taskName string, task recurring.Func)) {
		register("custom", func(context.Context, string) {})
		custom = true
	}

	server := &Server{Config: c, Storage: Storage{Tokens: allStorages().Tokens}, Hooks: h, scheduler: recurring.NewScheduler()}
	server.loadRecurringTasks()

	if !custom {
		t.Error("the functions of the custom tasks are not set")
	}
	for name, want := range map[string]bool{
		recurring.DeleteExpiredTokens: true,  // The storage is defined.
		recurring.SanitizeAccounts:    false, // The storage is not defined.
		"custom":                      true,  // The function is set by SetFuncOfRecurringTasks.
		"unknown":                     false,
	} {
		if got := server.scheduler.HasFunc(name); got != want {
			t.Errorf("HasFunc(%s) = %v, want %v", name, got, want)
		}
	}

	// Without a function to set the custom tasks.
	(&Server{Config: c, scheduler: recurring.NewScheduler()}).loadRecurringTasks()
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

	c := &config.Configuration{Environment: "test", Version: "1.0", TemplatesPath: filepath.Join(dir, "templates"), RunStartUpTasks: true}
	c.Hosts.Web = "https://web"
	c.SupportedLocales = []string{"en"}
	c.RecurringTasks = []config.RecurringTask{{Name: recurring.DeleteExpiredTokens}}

	var env, version string
	var calls []string
	s := &Server{Config: c, scheduler: recurring.NewScheduler(), Routes: weburl.NewConfiguration()}
	s.Routes.Routes["home"] = weburl.Route{AbsolutePaths: map[string]string{"en": "https://web/en/"}}
	s.Hooks.StartUpTasks = func(_ context.Context, environment, v string) {
		env, version = environment, v
		calls = append(calls, "start up")
	}
	s.Hooks.SetFuncOfRecurringTasks = func(func(string, recurring.Func)) { calls = append(calls, "recurring") }
	s.Storage.Tokens = allStorages().Tokens

	if err := s.setup(t.Context()); err != nil {
		t.Fatal(err)
	}

	if env != "test" || version != "1.0" || !slices.Equal(calls, []string{"start up", "recurring"}) {
		t.Errorf("env = %q, version = %q, calls = %v", env, version, calls)
	}
	if s.mailer == nil {
		t.Fatal("the mailer is not created")
	}
	if got := s.Routes.GetURL("home", "en"); got != "https://web/en/" {
		t.Errorf("URL of the home = %q", got)
	}
	if !s.scheduler.HasFunc(recurring.DeleteExpiredTokens) {
		t.Error("the recurring tasks are not loaded")
	}

	// The start up tasks only run if the configuration asks for it.
	c.RunStartUpTasks = false
	calls = nil
	if err := s.setup(t.Context()); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(calls, "start up") {
		t.Errorf("calls = %v", calls)
	}
}

func TestServer_Run_SetUpErrors(t *testing.T) {
	dir := t.TempDir()
	routes := filepath.Join(dir, "routes.yml")
	if err := os.WriteFile(routes, []byte("routes: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		routesFile string
		want       string
	}{
		{"missing templates", routes, "templates"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.Configuration{TemplatesPath: filepath.Join(dir, "templates")}
			s := &Server{Config: c, scheduler: recurring.NewScheduler()}

			if err := s.Run(t.Context()); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Run() = %v, want an error about %q", err, tt.want)
			}
		})
	}
}

func TestServer_Run_StopsWithTheContext(t *testing.T) {
	dir := t.TempDir()
	for path, content := range map[string]string{
		"templates/html/mails/hello.html": "<p>[[.Name]]</p>",
		"templates/text/mails/hello.txt":  "Hello {{.Name}}",
	} {
		path = filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	c := &config.Configuration{TemplatesPath: filepath.Join(dir, "templates")}
	c.RecurringTasks = []config.RecurringTask{{Name: "custom", Interval: 1, StartAt: "bad", Enabled: true}}
	s := &Server{Config: c, scheduler: recurring.NewScheduler()}
	s.Hooks.SetFuncOfRecurringTasks = func(register func(string, recurring.Func)) {
		register("custom", func(context.Context, string) {})
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() = %v, want nil once the context is done", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return once the context is done")
	}
}
