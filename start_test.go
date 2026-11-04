package ginamite

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/log"
)

func TestServer_ConnectPlugins(t *testing.T) {
	var events []string
	s := NewServer()
	s.AddPlugin(&testConnector{name: "first", events: &events})
	s.AddPlugin(&testPlugin{})
	s.AddPlugin(&testConnector{name: "second", events: &events})

	configuration := config.NewConfiguration()
	configuration.Database.Service = "db"
	closeAll, err := s.connectPlugins(configuration)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{"connect first db", "connect second db"}; !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}

	closeAll()

	if want := []string{"connect first db", "connect second db", "close second", "close first"}; !slices.Equal(events, want) {
		t.Errorf("events = %v, want %v", events, want)
	}
}

func TestServer_ConnectPlugins_NoConnector(t *testing.T) {
	s := NewServer()
	s.AddPlugin(&testPlugin{})

	closeAll, err := s.connectPlugins(config.NewConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	closeAll()
}

func TestServer_ConnectPlugins_Error(t *testing.T) {
	boom := errors.New("boom")

	var events []string
	s := NewServer()
	s.AddPlugin(&testConnector{name: "first", events: &events})
	s.AddPlugin(&testConnector{name: "second", events: &events, connectErr: boom})
	s.AddPlugin(&testConnector{name: "third", events: &events})

	closeAll, err := s.connectPlugins(config.NewConfiguration())

	if !errors.Is(err, boom) || closeAll != nil {
		t.Errorf("connectPlugins() = %v, want the error", err)
	}
	if want := []string{"connect first ", "connect second ", "close first"}; !slices.Equal(events, want) {
		t.Errorf("events = %v, want %v", events, want)
	}
}

func TestServer_SpecializedServer(t *testing.T) {
	tests := []struct {
		name  string
		flags flags.Flags
		want  func(s *Server) runner
	}{
		{"api", flags.Flags{IsAPIServer: true}, func(s *Server) runner { return s.API }},
		{"worker", flags.Flags{IsWorkerServer: true, WebRoutesFile: "routes.yaml"}, func(s *Server) runner { return s.Worker }},
		{"web", flags.Flags{IsWebServer: true, WebRoutesFile: "routes.yaml"}, func(s *Server) runner { return s.Web }},
		{"api wins", flags.Flags{IsAPIServer: true, IsWorkerServer: true, IsWebServer: true}, func(s *Server) runner { return s.API }},
		{"worker wins over web", flags.Flags{IsWorkerServer: true, IsWebServer: true}, func(s *Server) runner { return s.Worker }},
		{"none", flags.Flags{}, func(*Server) runner { return nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer()

			got := s.specializedServer(&tt.flags)

			if want := tt.want(s); got != want {
				t.Errorf("specializedServer() = %v, want %v", got, want)
			}
			if tt.name == "web" && s.Web.RoutesFile != "routes.yaml" {
				t.Errorf("Web.RoutesFile = %q, want the web routes file", s.Web.RoutesFile)
			}
		})
	}
}

func TestServer_InitializeConfiguration(t *testing.T) {
	dir := t.TempDir()
	locales := filepath.Join(dir, "locales")
	writeFile(t, filepath.Join(locales, "en-us.all.json"), `{"hello":"Hello"}`)
	conf := writeFile(t, filepath.Join(dir, "server.yaml"), "localization_path: "+locales+"\nlocales:\n  en-us:\n    default: true\n")

	oldLog := log.Main
	t.Cleanup(func() { log.Main = oldLog })

	var appFile string
	s := NewServer()
	s.ApplicationConfiguration.InitializeFunc = func(path string) { appFile = path }

	if err := s.initializeConfiguration(&flags.Flags{ServerName: "web", Environment: "test", Version: "final", ServerConfigurationFile: conf, ApplicationConfigurationFile: "app.yaml"}); err != nil {
		t.Fatal(err)
	}

	if s.Config == nil || s.Config.Environment != "test" {
		t.Fatalf("config = %+v", s.Config)
	}
	if s.API.Config != s.Config || s.Worker.Config != s.Config || s.Web.Config != s.Config {
		t.Error("the configuration must be shared with the specialized servers")
	}
	if appFile != "app.yaml" {
		t.Errorf("application file = %q", appFile)
	}
	if got := s.Localization.Translate("en-us")("hello"); got != "Hello" {
		t.Errorf("translation = %q", got)
	}
	if s.API.Localization != s.Localization || s.Worker.Localization != s.Localization || s.Web.Localization != s.Localization {
		t.Error("the translations must be shared with the specialized servers")
	}
}

func writeFile(t *testing.T, path, content string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

// fakeCommander is a plugin which handles the command "run".
type fakeCommander struct{ err error }

func (fakeCommander) Configure(*Server) error { return nil }

func (c fakeCommander) RegisterFlags(set *flag.FlagSet) {
	set.Bool("fake-run", false, "Run the fake command.")
}

func (c fakeCommander) Command(_ context.Context, f *flags.Flags, out io.Writer) (bool, error) {
	if f.ServerName != "run" {
		return false, nil
	}
	_, _ = io.WriteString(out, "ran")

	return true, c.err
}

func TestServer_runCommand(t *testing.T) {
	boom := errors.New("boom")

	s := NewServer()
	s.AddPlugin(&testPlugin{})
	s.AddPlugin(fakeCommander{err: boom})

	var out bytes.Buffer
	if handled, err := s.runCommand(t.Context(), &flags.Flags{ServerName: "run"}, &out); !handled || !errors.Is(err, boom) || out.String() != "ran" {
		t.Errorf("runCommand() = %v, %v, output %q", handled, err, out.String())
	}

	if handled, err := s.runCommand(t.Context(), &flags.Flags{}, &out); handled || err != nil {
		t.Errorf("runCommand() = %v, %v, want a command nobody handles to be left to the caller", handled, err)
	}
}

func TestServer_InitializeConfiguration_Errors(t *testing.T) {
	dir := t.TempDir()
	withoutTranslations := writeFile(t, filepath.Join(dir, "server.yaml"), "localization_path: "+dir+"\nlocales:\n  en-us:\n    default: true\n")

	oldLog := log.Main
	t.Cleanup(func() { log.Main = oldLog })

	tests := []struct {
		name, file string
	}{
		{"missing configuration", filepath.Join(dir, "missing.yaml")},
		{"missing translations", withoutTranslations},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewServer().initializeConfiguration(&flags.Flags{ServerName: "api", Environment: "test", ServerConfigurationFile: tt.file}); err == nil {
				t.Error("initializeConfiguration() must return an error")
			}
		})
	}
}

func TestServer_run_Errors(t *testing.T) {
	boom := errors.New("boom")

	dir := t.TempDir()
	conf := writeFile(t, filepath.Join(dir, "server.yaml"), "timezones_file_path: "+filepath.Join(dir, "missing.tab")+"\n")
	valid := writeFile(t, filepath.Join(dir, "valid.yaml"), "timezones_file_path: data/zone1970.tab\n")

	oldLog := log.Main
	t.Cleanup(func() { log.Main = oldLog })

	tests := []struct {
		name    string
		plugins []Plugin
		flags   flags.Flags
		want    error
		message string
	}{
		{"no server", nil, flags.Flags{}, nil, "no server selected"},
		{"invalid configuration", nil, flags.Flags{IsAPIServer: true, ServerConfigurationFile: filepath.Join(dir, "missing.yaml")}, nil, "configuration"},
		{
			"connection failure",
			[]Plugin{&testConnector{events: &[]string{}, connectErr: boom}},
			flags.Flags{IsAPIServer: true, ServerConfigurationFile: conf},
			boom, "",
		},
		{"missing timezones", nil, flags.Flags{IsWorkerServer: true, ServerConfigurationFile: conf}, nil, "timezones"},
		{"unreadable web routes", nil, flags.Flags{IsAPIServer: true, ServerConfigurationFile: valid, WebRoutesFile: filepath.Join(dir, "missing-routes.yml")}, nil, "routes"},
		{"missing robots", nil, flags.Flags{IsWebServer: true, ServerConfigurationFile: conf}, nil, "robots"},
		{"configuration failure", []Plugin{&testPlugin{err: boom}}, flags.Flags{IsAPIServer: true, ServerConfigurationFile: valid}, boom, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer()
			for _, p := range tt.plugins {
				s.AddPlugin(p)
			}

			tt.flags.ServerName, tt.flags.Environment = "test", "test"
			err := s.run(t.Context(), &tt.flags)

			switch {
			case err == nil:
				t.Error("run() must return an error")
			case tt.want != nil && !errors.Is(err, tt.want):
				t.Errorf("run() = %v, want %v", err, tt.want)
			case tt.message != "" && !strings.Contains(err.Error(), tt.message):
				t.Errorf("run() = %v, want an error about %q", err, tt.message)
			}
		})
	}
}

// withArgs replaces the arguments of the process for the duration of the test.
func withArgs(t *testing.T, args ...string) {
	t.Helper()

	old := os.Args
	os.Args = append([]string{"cmd"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func TestServer_Run_Arguments(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		args    []string
		plugins []Plugin
		want    error
		message string
	}{
		{"help", []string{"-help"}, nil, nil, ""},
		{"unknown flag", []string{"-unknown"}, nil, nil, "flag provided but not defined"},
		{"flag of a plugin", []string{"-fake-run", "-name", "run"}, []Plugin{fakeCommander{err: boom}}, boom, ""},
		{"flag of a missing plugin", []string{"-fake-run"}, nil, nil, "flag provided but not defined"},
		{"no server", []string{"-name", "n", "-conf-server", "s"}, nil, nil, "'api', 'worker' or 'web'"},
		{"application configuration", []string{"-api", "-name", "n", "-conf-server", "s"}, nil, nil, "application configuration file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withArgs(t, tt.args...)

			s := NewServer()
			s.ApplicationConfiguration.FileRequired = tt.name == "application configuration"
			for _, p := range tt.plugins {
				s.AddPlugin(p)
			}

			err := s.Run(t.Context())

			switch {
			case tt.want != nil && !errors.Is(err, tt.want):
				t.Errorf("Run() = %v, want %v", err, tt.want)
			case tt.message != "" && (err == nil || !strings.Contains(err.Error(), tt.message)):
				t.Errorf("Run() = %v, want an error about %q", err, tt.message)
			case tt.want == nil && tt.message == "" && err != nil:
				t.Errorf("Run() = %v", err)
			}
		})
	}
}

func TestServer_loadWebRoutes(t *testing.T) {
	file := writeFile(t, filepath.Join(t.TempDir(), "routes.yml"), "routes:\n  home:\n    paths:\n      en: /en/\n")

	t.Run("shared with the servers and the plugins", func(t *testing.T) {
		s := NewServer()
		s.Config.Hosts.Web = "https://web"

		if err := s.loadWebRoutes(&flags.Flags{WebRoutesFile: file}); err != nil {
			t.Fatal(err)
		}

		routes := s.WebRoutes()
		if routes == nil || routes.GetURL("home", "en") != "https://web/en/" {
			t.Fatalf("WebRoutes() = %+v", routes)
		}
		if s.API.WebRoutes != routes || s.Worker.Routes != routes || s.Web.Routes != routes {
			t.Error("the routes must be shared with the three servers")
		}
	})

	t.Run("no file", func(t *testing.T) {
		s := NewServer()
		var logs bytes.Buffer
		oldLog := log.Main
		t.Cleanup(func() { log.Main = oldLog })
		log.Main = slog.New(slog.NewTextHandler(&logs, nil))

		if err := s.loadWebRoutes(&flags.Flags{IsAPIServer: true}); err != nil {
			t.Fatal(err)
		}
		if out := logs.String(); !strings.Contains(out, "level=WARN") || !strings.Contains(out, "-web-routes") {
			t.Errorf("the absence of routes must be logged as a warning, logs = %q", out)
		}
		if s.WebRoutes() != nil || s.API.WebRoutes != nil || s.Worker.Routes != nil || s.Web.Routes != nil {
			t.Error("the routes must be nil without file")
		}
	})

	t.Run("unreadable file", func(t *testing.T) {
		s := NewServer()

		if err := s.loadWebRoutes(&flags.Flags{WebRoutesFile: file + ".missing"}); err == nil {
			t.Error("loadWebRoutes() must fail")
		}
	})
}
