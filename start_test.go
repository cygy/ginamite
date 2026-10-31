package ginamite

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/localization"
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
		{"api", flags.Flags{IsAPIServer: true, RoutesFile: "routes.yaml"}, func(s *Server) SpecializedServer { return s.API }},
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
				if routes := map[string]string{"api": s.API.RoutesFile, "worker": s.Worker.RoutesFile, "web": s.Web.RoutesFile}; routes[tt.name] != tt.flags.RoutesFile && tt.flags.RoutesFile != "" {
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

	oldLog, oldLocalization := log.Main, localization.Main
	t.Cleanup(func() { log.Main, localization.Main = oldLog, oldLocalization })

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
	if got := localization.Translate("en-us")("hello"); got != "Hello" {
		t.Errorf("translation = %q", got)
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
