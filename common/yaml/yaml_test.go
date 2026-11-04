package yaml_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/yaml"
)

type settings struct {
	Secret   string   `yaml:"secret"`
	Password string   `yaml:"password"`
	Port     int      `yaml:"port"`
	Plain    string   `yaml:"plain"`
	Keys     []string `yaml:"keys"`
}

func TestUnmarshal(t *testing.T) {
	file := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(file, []byte("from-a-file\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_SECRETS_PASSWORD", "from-env")
	t.Setenv("TEST_SECRETS_PORT", "8085")

	var s settings
	source := "secret: !file " + file + "\npassword: !env TEST_SECRETS_PASSWORD\nport: !env TEST_SECRETS_PORT\nplain: kept\nkeys: [a, !env TEST_SECRETS_PASSWORD]\n"
	if err := yaml.Unmarshal([]byte(source), &s); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if s.Secret != "from-a-file" || s.Password != "from-env" || s.Port != 8085 || s.Plain != "kept" {
		t.Errorf("settings = %+v", s)
	}
	if len(s.Keys) != 2 || s.Keys[1] != "from-env" {
		t.Errorf("keys = %v", s.Keys)
	}
}

func TestUnmarshal_DecodesEveryOut(t *testing.T) {
	t.Setenv("TEST_SECRETS_PASSWORD", "from-env")

	var a, b settings
	if err := yaml.Unmarshal([]byte("password: !env TEST_SECRETS_PASSWORD"), &a, &b); err != nil {
		t.Fatal(err)
	}
	if a.Password != "from-env" || b.Password != "from-env" {
		t.Errorf("a = %+v, b = %+v", a, b)
	}
}

func TestUnmarshal_Errors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	tests := []struct{ name, source, want string }{
		{"missing file", "password: !file " + missing, "password: unable to read the file " + missing},
		{"unset variable", "secret: !env TEST_SECRETS_UNSET", "secret: the environment variable TEST_SECRETS_UNSET is not set"},
		{"no value", "secret: !env", "secret: the tag !env needs a value"},
		{"in a list", "keys:\n  - !env TEST_SECRETS_UNSET", "keys[0]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := yaml.Unmarshal([]byte(tt.source), &settings{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestUnmarshal_EmptyFile(t *testing.T) {
	if err := yaml.Unmarshal(nil, &settings{}); err != nil {
		t.Errorf("Unmarshal() error = %v", err)
	}
}
