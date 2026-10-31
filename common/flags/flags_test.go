package flags

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestNewFlags(t *testing.T) {
	if f := NewFlags(); f == nil || *f != (Flags{}) {
		t.Errorf("NewFlags() = %+v, want an empty Flags", f)
	}
}

func TestFlags_parse(t *testing.T) {
	valid := []string{"-api", "-name", "srv", "-conf-server", "s.yml", "-routes", "r.yml"}

	tests := []struct {
		name         string
		args         []string
		appConfig    bool
		wantOK       bool
		wantProblem  string
		wantUsage    bool
		checkSuccess func(t *testing.T, f *Flags)
	}{
		{
			name:   "valid api server",
			args:   valid,
			wantOK: true,
			checkSuccess: func(t *testing.T, f *Flags) {
				t.Helper()
				if !f.IsAPIServer || f.IsWorkerServer || f.IsWebServer || f.ServerName != "srv" ||
					f.ServerConfigurationFile != "s.yml" || f.RoutesFile != "r.yml" ||
					f.Environment != "development" || f.Version != "final" {
					t.Errorf("unexpected flags %+v", f)
				}
			},
		},
		{
			name:      "valid with application configuration",
			args:      append([]string{"-worker", "-env", "production", "-version", "beta", "-conf-app", "a.yml"}, valid[1:]...),
			appConfig: true,
			wantOK:    true,
			checkSuccess: func(t *testing.T, f *Flags) {
				t.Helper()
				if !f.IsWorkerServer || f.Environment != "production" || f.Version != "beta" || f.ApplicationConfigurationFile != "a.yml" {
					t.Errorf("unexpected flags %+v", f)
				}
			},
		},
		{name: "help", args: []string{"-help"}, wantUsage: true},
		{name: "unknown flag", args: []string{"-unknown"}, wantProblem: "flag provided but not defined"},
		{name: "no server type", args: valid[1:], wantProblem: "'api', 'worker' or 'web'"},
		{name: "no server name", args: []string{"-web", "-conf-server", "s", "-routes", "r"}, wantProblem: "server name"},
		{name: "no server configuration", args: []string{"-web", "-name", "n", "-routes", "r"}, wantProblem: "server configuration file"},
		{name: "no application configuration", args: valid, appConfig: true, wantProblem: "application configuration file"},
		{name: "no routes", args: []string{"-web", "-name", "n", "-conf-server", "s"}, wantProblem: "routes file"},
		{
			name:   "dead letters command, which needs no server nor routes",
			args:   []string{"-dlq", "replay", "-dlq-topic", "mail", "-dlq-limit", "5", "-dlq-body", "-name", "n", "-conf-server", "s.yml"},
			wantOK: true,
			checkSuccess: func(t *testing.T, f *Flags) {
				t.Helper()
				if f.DLQ != DLQReplay || f.DLQTopic != "mail" || f.DLQLimit != 5 || !f.DLQBody || f.IsAPIServer || f.RoutesFile != "" {
					t.Errorf("unexpected flags %+v", f)
				}
			},
		},
		{name: "unknown dead letters command", args: []string{"-dlq", "delete", "-dlq-topic", "mail", "-name", "n", "-conf-server", "s"}, wantProblem: "-dlq must be 'list', 'replay' or 'purge'"},
		{name: "dead letters command with a server", args: []string{"-dlq", "list", "-api", "-dlq-topic", "mail", "-name", "n", "-conf-server", "s"}, wantProblem: "starts no server"},
		{name: "dead letters command without topic", args: []string{"-dlq", "list", "-name", "n", "-conf-server", "s"}, wantProblem: "-dlq-topic"},
		{name: "dead letters command with a negative limit", args: []string{"-dlq", "list", "-dlq-topic", "mail", "-dlq-limit", "-1", "-name", "n", "-conf-server", "s"}, wantProblem: "must not be negative"},
		{name: "dead letters command without name", args: []string{"-dlq", "list", "-dlq-topic", "mail", "-conf-server", "s"}, wantProblem: "server name"},
		{name: "dead letters command without configuration", args: []string{"-dlq", "list", "-dlq-topic", "mail", "-name", "n"}, wantProblem: "server configuration file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			f := NewFlags()

			ok, problem := f.parse(tt.args, &output, tt.appConfig)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantProblem == "" && problem != "" || !strings.Contains(problem, tt.wantProblem) {
				t.Errorf("problem = %q, want it to contain %q", problem, tt.wantProblem)
			}
			if tt.wantUsage && !strings.Contains(output.String(), "-conf-server") {
				t.Errorf("the usage is not printed: %q", output.String())
			}
			if tt.checkSuccess != nil {
				tt.checkSuccess(t, f)
			}
		})
	}
}

func TestFlags_Parse(t *testing.T) {
	defer func(old []string) { os.Args = old }(os.Args)

	os.Args = []string{"cmd", "-web", "-name", "srv", "-conf-server", "s.yml", "-routes", "r.yml"}

	f := NewFlags()
	if !f.Parse(false) || !f.IsWebServer || f.ServerName != "srv" {
		t.Errorf("Parse() failed: %+v", f)
	}

	os.Args = []string{"cmd", "-help"}
	if NewFlags().Parse(false) {
		t.Error("Parse() must return false when the help is requested")
	}
}
