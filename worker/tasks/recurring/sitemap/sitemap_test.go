package sitemap

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

// setup sets the configuration and the routes used by Generate, for the duration of the test, and returns
// the directory of the sitemaps.
func setup(t *testing.T, maxURLsPerPage int, routes map[string]weburl.Route) string {
	t.Helper()

	oldConfig, oldRoutes := config.Main, weburl.Main
	t.Cleanup(func() { config.Main, weburl.Main = oldConfig, oldRoutes })

	dir := t.TempDir()
	config.Main = &config.Configuration{SupportedLocales: []string{"en", "fr"}}
	config.Main.Sitemap = config.SitemapConfiguration{Path: dir, Name: "sitemap", MaximumURLsPerPage: maxURLsPerPage}
	config.Main.Hosts.Static = "https://static"
	weburl.Main = &weburl.Configuration{Routes: routes}

	return dir
}

// route returns a route included in the sitemap, with the same path in each locale.
func route(path string) weburl.Route {
	return weburl.Route{
		RelativePaths: map[string]string{"en": path, "fr": path},
		AbsolutePaths: map[string]string{"en": "https://web/en" + path, "fr": "https://web/fr" + path},
		Sitemap:       html.Sitemap{Included: true, Priority: 0.5, UpdateFrequency: "weekly"},
	}
}

func TestGenerate(t *testing.T) {
	workertest.Hook.Reset()

	notIncluded := route("/hidden")
	notIncluded.Sitemap.Included = false
	onlyEnglish := route("/only")
	delete(onlyEnglish.AbsolutePaths, "fr")

	dir := setup(t, 1000, map[string]weburl.Route{
		"home":   route("/home"),
		"hidden": notIncluded,
		"only":   onlyEnglish,
		"page":   route("/page/:id"),
	})

	var requests []string
	Generate("generate", func(key string, offset, limit int) map[string][]map[string]string {
		requests = append(requests, fmt.Sprintf("%s:%d:%d", key, offset, limit))

		// Three batches: the maximum of 1000 URLs per file is reached at the second one.
		if offset >= 3*batchSize {
			return nil
		}

		return map[string][]map[string]string{
			"en":      {{"id": fmt.Sprint(offset)}},
			"fr":      {{"id": fmt.Sprint(offset)}},
			"unknown": {{"id": "x"}}, // Neither a supported locale nor a locale of the route.
		}
	})

	if got, want := strings.Join(requests, " "), "page:0:500 page:500:500 page:1000:500 page:1500:500"; got != want {
		t.Errorf("requests = %q, want %q", got, want)
	}

	wantFiles := []string{
		"sitemap_en_main.xml.gz", "sitemap_fr_main.xml.gz", "sitemap_index.xml.gz",
		"sitemap_en_page_0.xml.gz", "sitemap_en_page_1.xml.gz", "sitemap_fr_page_0.xml.gz", "sitemap_fr_page_1.xml.gz",
	}
	got := files(t, dir)
	if len(got) != len(wantFiles) {
		t.Fatalf("files = %v, want %v", got, wantFiles)
	}

	read := func(name string) string { return readGzip(t, filepath.Join(dir, name)) }

	// The main sitemaps.
	en := read("sitemap_en_main.xml.gz")
	if !strings.Contains(en, "<loc>https://web/en/home</loc>") || !strings.Contains(en, "<loc>https://web/en/only</loc>") ||
		!strings.Contains(en, "<changefreq>weekly</changefreq>") || strings.Contains(en, "hidden") || strings.Contains(en, "page") {
		t.Errorf("main sitemap = %q", en)
	}
	fr := read("sitemap_fr_main.xml.gz")
	if !strings.Contains(fr, "<loc>https://web/fr/home</loc>") || strings.Contains(fr, "only") {
		t.Errorf("main sitemap = %q", fr)
	}

	// The sitemaps of the route with variables: one URL in the first file and two in the second one.
	for locale, prefix := range map[string]string{"en": "https://web/en/page/", "fr": "https://web/fr/page/"} {
		first := read(fmt.Sprintf("sitemap_%s_page_0.xml.gz", locale))
		second := read(fmt.Sprintf("sitemap_%s_page_1.xml.gz", locale))

		if strings.Count(first, "<url>") != 1 || !strings.Contains(first, prefix+"0<") {
			t.Errorf("first sitemap = %q", first)
		}
		if strings.Count(second, "<url>") != 2 || !strings.Contains(second, prefix+"500<") || !strings.Contains(second, prefix+"1000<") {
			t.Errorf("second sitemap = %q", second)
		}
	}

	// The index references each sitemap.
	index := read("sitemap_index.xml.gz")
	for _, name := range wantFiles {
		if name != "sitemap_index.xml.gz" && !strings.Contains(index, "<loc>https://static"+filepath.Join(dir, name)+"</loc>") {
			t.Errorf("the index does not reference %s: %q", name, index)
		}
	}

	if entry := workertest.Hook.LastEntry(); entry == nil || entry.Message != "task done" {
		t.Errorf("last log = %+v", entry)
	}
}

func TestGenerate_NoMaximum(t *testing.T) {
	dir := setup(t, 0, map[string]weburl.Route{"page": route("/page/:id")})

	Generate("generate", func(_ string, offset, _ int) map[string][]map[string]string {
		if offset >= 4*batchSize {
			return nil
		}

		return map[string][]map[string]string{"en": {{"id": "x"}}}
	})

	if content := readGzip(t, filepath.Join(dir, "sitemap_en_page_0.xml.gz")); strings.Count(content, "<url>") != 4 {
		t.Errorf("sitemap = %q, want 4 URLs in a single file", content)
	}
	if names := files(t, dir); len(names) != 5 { // the index, 2 main sitemaps and 2 sitemaps of the route.
		t.Errorf("files = %v", names)
	}
}

func TestGenerate_Errors(t *testing.T) {
	t.Run("no index", func(t *testing.T) {
		workertest.Hook.Reset()
		dir := setup(t, 1000, map[string]weburl.Route{"home": route("/home")})
		config.Main.Sitemap.Path = filepath.Join(dir, "missing")

		Generate("generate", nil)

		entry := workertest.Hook.LastEntry()
		if entry == nil || entry.Message != "task error" || entry.Data["filename"] != "sitemap_index" {
			t.Errorf("last log = %+v", entry)
		}
	})

	t.Run("sitemaps which can not be created or closed", func(t *testing.T) {
		workertest.Hook.Reset()
		dir := setup(t, 1000, map[string]weburl.Route{"home": route("/home"), "page": route("/page/:id")})

		// A directory in place of the main sitemap of a locale prevents the move of the temporary file.
		blocked := filepath.Join(dir, "sitemap_en_main.xml.gz", "keep")
		if err := mkdirAll(blocked); err != nil {
			t.Fatal(err)
		}

		Generate("generate", func(string, int, int) map[string][]map[string]string { return nil })

		errors := 0
		for _, entry := range workertest.Hook.AllEntries() {
			if entry.Message == "task error" && entry.Data["filename"] == "sitemap_en_main.xml.gz" {
				errors++
			}
		}
		if errors != 1 {
			t.Errorf("logged %d errors for the main sitemap, want 1", errors)
		}

		index := readGzip(t, filepath.Join(dir, "sitemap_index.xml.gz"))
		if strings.Contains(index, "sitemap_en_main") || !strings.Contains(index, "sitemap_fr_main") {
			t.Errorf("index = %q, want only the sitemaps closed without error", index)
		}
	})
}

func TestGenerator_newFile_Error(t *testing.T) {
	workertest.Hook.Reset()

	g := generator{taskName: "generate", name: "sitemap", path: filepath.Join(t.TempDir(), "missing"), locales: []string{"en", "fr"}}

	if f, ok := g.newFile("sitemap_en_main", false); ok || f != nil {
		t.Errorf("file = %v, ok = %v", f, ok)
	}
	if entry := workertest.Hook.LastEntry(); entry == nil || entry.Message != "task error" || entry.Data["filename"] != "sitemap_en_main" {
		t.Errorf("last log = %+v", entry)
	}

	if sitemaps := g.createSitemaps("page", 0); len(sitemaps) != 0 {
		t.Errorf("sitemaps = %v, want none", sitemaps)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		variables map[string]string
		want      string
	}{
		{"no variable", "/page", nil, "/page"},
		{"one variable", "/page/:id", map[string]string{"id": "42"}, "/page/42"},
		{"several variables", "/:category/:id", map[string]string{"id": "42", "category": "books"}, "/books/42"},
		{"a name is the start of another", "/:id/:identifier", map[string]string{"id": "1", "identifier": "abc"}, "/1/abc"},
		{"unknown variable", "/:id", map[string]string{"other": "x"}, "/:id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The result must not depend on the order of the map.
			for range 20 {
				if got := expandPath(tt.path, tt.variables); got != tt.want {
					t.Fatalf("expandPath = %q, want %q", got, tt.want)
				}
			}
		})
	}
}
