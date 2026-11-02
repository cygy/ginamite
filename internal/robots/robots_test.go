package robots

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "robots.txt")
	if err := os.WriteFile(path, []byte("User-agent: *\nDisallow: /private\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	content, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "User-agent: *\nDisallow: /private\n"; content != want {
		t.Errorf("content = %q, want %q", content, want)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	content, err := Load(filepath.Join(t.TempDir(), "missing.txt"))

	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error = %v, want one wrapping fs.ErrNotExist", err)
	}
	if content != "" {
		t.Errorf("content = %q, want it empty", content)
	}
}
