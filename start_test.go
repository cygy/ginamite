package ginamite

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
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
	closeAll := s.connectPlugins(configuration)

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

	s.connectPlugins(config.NewConfiguration())()
}

func TestServer_SpecializedServer(t *testing.T) {
	tests := []struct {
		name  string
		flags flags.Flags
		want  func(s *Server) SpecializedServer
	}{
		{"api", flags.Flags{IsAPIServer: true}, func(s *Server) SpecializedServer { return s.API }},
		{"worker", flags.Flags{IsWorkerServer: true, RoutesFile: "routes.yaml"}, func(s *Server) SpecializedServer { return s.Worker }},
		{"web", flags.Flags{IsWebServer: true, RoutesFile: "routes.yaml"}, func(s *Server) SpecializedServer { return s.Web }},
		{"api wins", flags.Flags{IsAPIServer: true, IsWorkerServer: true, IsWebServer: true}, func(s *Server) SpecializedServer { return s.API }},
		{"worker wins over web", flags.Flags{IsWorkerServer: true, IsWebServer: true}, func(s *Server) SpecializedServer { return s.Worker }},
		{"none", flags.Flags{}, func(*Server) SpecializedServer { return nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer()

			got := s.specializedServer(&tt.flags)

			if want := tt.want(s); got != want {
				t.Errorf("specializedServer() = %v, want %v", got, want)
			}
			if got != nil {
				if routes := map[string]string{"worker": s.Worker.RoutesFile, "web": s.Web.RoutesFile}; routes[tt.name] != tt.flags.RoutesFile && tt.flags.RoutesFile != "" {
					t.Errorf("routes files = %v", routes)
				}
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

	s.initializeConfiguration(&flags.Flags{ServerName: "web", Environment: "test", Version: "final", ServerConfigurationFile: conf, ApplicationConfigurationFile: "app.yaml"})

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

func (fakeCommander) Configure(*Server) {}

func (c fakeCommander) Command(f *flags.Flags, out io.Writer) (bool, error) {
	if f.DLQ != "run" {
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
	if handled, err := s.runCommand(&flags.Flags{DLQ: "run"}, &out); !handled || !errors.Is(err, boom) || out.String() != "ran" {
		t.Errorf("runCommand() = %v, %v, output %q", handled, err, out.String())
	}

	if handled, err := s.runCommand(&flags.Flags{}, &out); handled || err != nil {
		t.Errorf("runCommand() = %v, %v, want a command nobody handles to be left to the caller", handled, err)
	}
}
