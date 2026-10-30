package robots

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestInitialize(t *testing.T) {
	old := Content
	t.Cleanup(func() { Content = old })

	path := webtest.WriteFile(t, t.TempDir(), "robots.txt", "User-agent: *\nDisallow: /private\n")

	if err := Initialize(path); err != nil {
		t.Fatal(err)
	}
	if want := "User-agent: *\nDisallow: /private\n"; Content != want {
		t.Errorf("Content = %q, want %q", Content, want)
	}
}

func TestInitialize_MissingFile(t *testing.T) {
	old := Content
	Content = "previous"
	t.Cleanup(func() { Content = old })

	path := filepath.Join(t.TempDir(), "missing.txt")

	err := Initialize(path)

	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error = %v, want one wrapping fs.ErrNotExist", err)
	}
	if Content != "previous" {
		t.Errorf("Content = %q, want it unchanged", Content)
	}
}
