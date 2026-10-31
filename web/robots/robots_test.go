package robots

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestLoad(t *testing.T) {
	path := webtest.WriteFile(t, t.TempDir(), "robots.txt", "User-agent: *\nDisallow: /private\n")

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
