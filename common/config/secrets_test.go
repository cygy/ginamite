package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"
)

func TestConfiguration_Initialize_ExternalValues(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "jwt_secret")
	if err := os.WriteFile(secret, []byte("a-secret-read-from-a-file-of-32-chars\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_DB_PASSWORD", "from-env")
	t.Setenv("TEST_PORT", "8085")

	file := writeFile(t, `
jwt:
  secret: !file `+secret+`
database:
  password: !env TEST_DB_PASSWORD
  host: plain
server:
  port: !env TEST_PORT
`)

	c := NewConfiguration()
	if err := c.Initialize(file, environment.Production, version.Final); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if c.JWT.Secret != "a-secret-read-from-a-file-of-32-chars" {
		t.Errorf("secret = %q, want the content of the file without its line break", c.JWT.Secret)
	}
	if c.Database.Password != "from-env" || c.Database.Host != "plain" {
		t.Errorf("database = %+v", c.Database)
	}
	if c.Server.Port != 8085 {
		t.Errorf("port = %d, want the integer read from the environment", c.Server.Port)
	}
}

func TestConfiguration_Initialize_ExternalValues_Errors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")

	tests := []struct {
		name string
		yaml string
		want string
	}{
		{"missing file", "database:\n  password: !file " + missing, "database.password: unable to read the file " + missing},
		{"unset variable", "jwt:\n  secret: !env TEST_UNSET_VARIABLE", "jwt.secret: the environment variable TEST_UNSET_VARIABLE is not set"},
		{"no value", "jwt:\n  secret: !env", "jwt.secret: the tag !env needs a value"},
		{"in a list", "worker_consumers:\n  - topic: !env TEST_UNSET_VARIABLE", "worker_consumers[0].topic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := unmarshal([]byte(tt.yaml), NewConfiguration())
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestUnmarshal_EmptyFile(t *testing.T) {
	if err := unmarshal(nil, NewConfiguration()); err != nil {
		t.Errorf("unmarshal() error = %v", err)
	}
}

func TestConfiguration_Initialize_RefusesThePlaceholderSecretInProduction(t *testing.T) {
	file := writeFile(t, "jwt:\n  secret: change-me-with-at-least-32-characters-secret\n")

	t.Run("production", func(t *testing.T) {
		if err := NewConfiguration().Initialize(file, environment.Production, version.Final); !errors.Is(err, errPlaceholderSecret) {
			t.Errorf("Initialize() error = %v, want errPlaceholderSecret", err)
		}
	})

	t.Run("development", func(t *testing.T) {
		if err := NewConfiguration().Initialize(file, environment.Development, version.Final); err != nil {
			t.Errorf("Initialize() error = %v", err)
		}
	})
}
