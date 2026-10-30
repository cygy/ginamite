package sitemap

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/worker/internal/workertest"
)

// readGzip returns the uncompressed content of a file.
func readGzip(t *testing.T, path string) string {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	r, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// files returns the names of the files of a directory.
func files(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names
}

func TestNewFile(t *testing.T) {
	t.Run("sitemap", func(t *testing.T) {
		dir := t.TempDir()

		f, err := NewFile(dir, "sitemap_en", false)
		if err != nil {
			t.Fatal(err)
		}
		if f.Name != "sitemap_en.xml.gz" || f.Path != filepath.Join(dir, "sitemap_en.xml.gz") {
			t.Errorf("Name = %q, Path = %q", f.Name, f.Path)
		}

		// The file only exists once closed.
		if names := files(t, dir); len(names) != 1 || names[0] == f.Name {
			t.Errorf("files = %v, want only the temporary file", names)
		}

		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		if names := files(t, dir); len(names) != 1 || names[0] != f.Name {
			t.Errorf("files = %v, want only %q", names, f.Name)
		}

		content := readGzip(t, f.Path)
		if !strings.Contains(content, "<urlset xmlns=") || !strings.HasSuffix(content, "</urlset>") {
			t.Errorf("content = %q", content)
		}
	})

	t.Run("index", func(t *testing.T) {
		f, err := NewFile(t.TempDir(), "index", true)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		content := readGzip(t, f.Path)
		if !strings.Contains(content, "<sitemapindex xmlns=") || !strings.HasSuffix(content, "</sitemapindex>") {
			t.Errorf("content = %q", content)
		}
	})

	t.Run("unwritable directory", func(t *testing.T) {
		workertest.Hook.Reset()

		if f, err := NewFile(filepath.Join(t.TempDir(), "missing"), "x", false); err == nil || f != nil {
			t.Errorf("file = %v, err = %v", f, err)
		}
		if entry := workertest.Hook.LastEntry(); entry == nil || entry.Message != "create sitemap file" {
			t.Errorf("log = %+v", entry)
		}
	})
}

func TestFile_AddURL(t *testing.T) {
	f, err := NewFile(t.TempDir(), "sitemap", false)
	if err != nil {
		t.Fatal(err)
	}

	f.AddURL("https://example.com/search?a=1&b=<2>", "daily", 0.55)
	f.AddSitemap("https://example.com/ignored.xml.gz", time.Now())
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	content := readGzip(t, f.Path)
	want := "<url>\n<loc>https://example.com/search?a=1&amp;b=&lt;2&gt;</loc>\n<changefreq>daily</changefreq>\n<priority>0.6</priority>\n</url>\n"
	if !strings.Contains(content, want) || strings.Contains(content, "ignored") {
		t.Errorf("content = %q, want it to contain %q only", content, want)
	}
}

func TestFile_AddSitemap(t *testing.T) {
	f, err := NewFile(t.TempDir(), "index", true)
	if err != nil {
		t.Fatal(err)
	}

	f.AddSitemap("https://example.com/a&b.xml.gz", time.Date(2026, 10, 30, 12, 0, 0, 0, time.FixedZone("CET", 3600)))
	f.AddURL("https://example.com/ignored", "daily", 0.5)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	content := readGzip(t, f.Path)
	want := "<sitemap>\n<loc>https://example.com/a&amp;b.xml.gz</loc>\n<lastmod>2026-10-30T12:00:00+01:00</lastmod>\n</sitemap>\n"
	if !strings.Contains(content, want) || strings.Contains(content, "ignored") {
		t.Errorf("content = %q, want it to contain %q only", content, want)
	}
}

func TestFile_Close_Errors(t *testing.T) {
	t.Run("the destination is not writable", func(t *testing.T) {
		workertest.Hook.Reset()
		dir := t.TempDir()

		f, err := NewFile(dir, "sitemap", false)
		if err != nil {
			t.Fatal(err)
		}
		f.Path = filepath.Join(dir, "missing", "sitemap.xml.gz")

		if err := f.Close(); err == nil {
			t.Error("expected an error")
		}
		if names := files(t, dir); len(names) != 0 {
			t.Errorf("files = %v, the temporary file must be deleted", names)
		}
		if entry := workertest.Hook.LastEntry(); entry == nil || entry.Message != "move temp sitemap file" {
			t.Errorf("log = %+v", entry)
		}
	})

	t.Run("the file is closed", func(t *testing.T) {
		workertest.Hook.Reset()
		dir := t.TempDir()

		f, err := NewFile(dir, "sitemap", false)
		if err != nil {
			t.Fatal(err)
		}
		_ = f.file.Close()

		if err := f.Close(); err == nil {
			t.Error("expected an error")
		}
		if names := files(t, dir); len(names) != 0 {
			t.Errorf("files = %v, nothing must be published", names)
		}

		messages := []string{}
		for _, entry := range workertest.Hook.AllEntries() {
			messages = append(messages, entry.Message)
		}
		for _, want := range []string{"flush to sitemap file", "close sitemap writer", "close sitemap file"} {
			found := false
			for _, message := range messages {
				found = found || message == want
			}
			if !found {
				t.Errorf("%q is not logged, got %v", want, messages)
			}
		}
	})
}

func TestFirstError(t *testing.T) {
	if err := firstError(nil, nil); err != nil {
		t.Errorf("err = %v", err)
	}

	first, second := os.ErrClosed, os.ErrExist
	if err := firstError(nil, first, second); err != first {
		t.Errorf("err = %v, want %v", err, first)
	}
}

// mkdirAll creates a directory and its parents.
func mkdirAll(path string) error {
	return os.MkdirAll(path, 0o750)
}
