package flags

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestNewFlags(t *testing.T) {
	if f := NewFlags(); f == nil || *f != (Flags{}) {
		t.Errorf("NewFlags() = %+v, want an empty Flags", f)
	}
}

func TestFlags_Parse(t *testing.T) {
	valid := []string{"-api", "-name", "srv", "-conf-server", "s.yml", "-routes", "r.yml"}

	t.Run("api server", func(t *testing.T) {
		f := NewFlags()

		if err := f.Parse(valid, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		if !f.IsAPIServer || f.IsWorkerServer || f.IsWebServer || f.ServerName != "srv" ||
			f.ServerConfigurationFile != "s.yml" || f.RoutesFile != "r.yml" ||
			f.Environment != "development" || f.Version != "final" {
			t.Errorf("unexpected flags %+v", f)
		}
	})

	t.Run("all the arguments", func(t *testing.T) {
		f := NewFlags()
		args := append([]string{"-worker", "-env", "production", "-version", "beta", "-conf-app", "a.yml"}, valid[1:]...)

		if err := f.Parse(args, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		if !f.IsWorkerServer || f.Environment != "production" || f.Version != "beta" || f.ApplicationConfigurationFile != "a.yml" {
			t.Errorf("unexpected flags %+v", f)
		}
	})

	t.Run("flags of the plugins", func(t *testing.T) {
		var topic string
		f := NewFlags()

		err := f.Parse([]string{"-name", "n", "-topic", "mail"}, &bytes.Buffer{}, func(set *flag.FlagSet) {
			set.StringVar(&topic, "topic", "", "A topic.")
		})

		if err != nil || topic != "mail" || f.ServerName != "n" {
			t.Errorf("Parse() = %v, topic = %q, flags = %+v", err, topic, f)
		}
	})

	t.Run("help", func(t *testing.T) {
		var output bytes.Buffer

		err := NewFlags().Parse([]string{"-help"}, &output, func(set *flag.FlagSet) {
			set.Bool("extra", false, "A flag of a plugin.")
		})

		if !errors.Is(err, flag.ErrHelp) {
			t.Errorf("Parse() = %v, want flag.ErrHelp", err)
		}
		if usage := output.String(); !strings.Contains(usage, "-conf-server") || !strings.Contains(usage, "-extra") {
			t.Errorf("the usage is not printed: %q", usage)
		}
	})

	t.Run("unknown flag", func(t *testing.T) {
		err := NewFlags().Parse([]string{"-unknown"}, &bytes.Buffer{})

		if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
			t.Errorf("Parse() = %v", err)
		}
	})
}

func TestFlags_Validate(t *testing.T) {
	tests := []struct {
		name      string
		flags     Flags
		appConfig bool
		want      string
	}{
		{"api server", Flags{IsAPIServer: true, ServerName: "n", ServerConfigurationFile: "s"}, false, ""},
		{"web server", Flags{IsWebServer: true, ServerName: "n", ServerConfigurationFile: "s", RoutesFile: "r"}, false, ""},
		{"with application configuration", Flags{IsWorkerServer: true, ServerName: "n", ServerConfigurationFile: "s", ApplicationConfigurationFile: "a"}, true, ""},
		{"no server type", Flags{ServerName: "n", ServerConfigurationFile: "s"}, false, "'api', 'worker' or 'web'"},
		{"no server name", Flags{IsAPIServer: true, ServerConfigurationFile: "s"}, false, "server name"},
		{"no server configuration", Flags{IsAPIServer: true, ServerName: "n"}, false, "server configuration file"},
		{"no application configuration", Flags{IsAPIServer: true, ServerName: "n", ServerConfigurationFile: "s"}, true, "application configuration file"},
		{"no routes", Flags{IsWebServer: true, ServerName: "n", ServerConfigurationFile: "s"}, false, "routes file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flags.Validate(tt.appConfig)

			switch {
			case tt.want == "" && err != nil:
				t.Errorf("Validate() = %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Errorf("Validate() = %v, want an error about %q", err, tt.want)
			}
		})
	}
}
