package html

import (
	"reflect"
	"testing"

	"github.com/cygy/ginamite/common/weburl"
)

func TestPageCopy(t *testing.T) {
	// Create a new page struct.
	page := generatePage()

	// Copy the page struct.
	copiedPage := page.Copy()

	// Check the fields.
	if page.Locale != copiedPage.Locale {
		t.Error("To copy a Page struct must not modified the property 'Locale'.")
	}

	if page.Timezone != copiedPage.Timezone {
		t.Error("To copy a Page struct must not modified the property 'Timezone'.")
	}

	if page.Title != copiedPage.Title {
		t.Error("To copy a Page struct must not modified the property 'Title'.")
	}

	if !reflect.DeepEqual(page.Paths, copiedPage.Paths) {
		t.Error("To copy a Page struct must not modified the property 'Paths'.")
	}

	if page.Template != copiedPage.Template {
		t.Error("To copy a Page struct must not modified the property 'Template'.")
	}

	if page.AccessRight != copiedPage.AccessRight {
		t.Error("To copy a Page struct must not modified the property 'AccessRight'.")
	}

	if !reflect.DeepEqual(page.ScriptFiles, copiedPage.ScriptFiles) {
		t.Error("To copy a Page struct must not modified the property 'ScriptFiles'.")
	}

	if len(copiedPage.Meta.Title) == 0 {
		t.Error("To copy a Page struct must not modified the property 'Meta'.")
	}

	if copiedPage.Sitemap.Priority != 1.0 {
		t.Error("To copy a Page struct must not modified the property 'Priority'.")
	}

	if page.Partial != copiedPage.Partial {
		t.Error("To copy a Page struct must not modified the property 'Partial'.")
	}

	if page.Key != copiedPage.Key {
		t.Error("To copy a Page struct must not modified the property 'Key'.")
	}
}

func TestPageMerge(t *testing.T) {
	// Create a new page struct.
	page := generatePage()

	originalScriptFiles := page.ScriptFiles

	// Create another page struct.
	pageToMerge := generatePage()
	pageToMerge.Locale = "en-us"
	pageToMerge.Timezone = "Europe/Paris"
	pageToMerge.Title = "title_2"
	pageToMerge.Paths = map[string]string{
		"fr-fr": "/path-ca",
		"en-gb": "/path-en",
	}
	pageToMerge.Template = PageTemplate{
		Decorator: "template_decorator_2",
		Main:      "template_main_2",
	}
	pageToMerge.AccessRight = AccessRightAll
	pageToMerge.ScriptFiles = []Script{
		{
			Src: "script_3.js",
		},
		{
			Src: "script_4.js",
		},
	}
	pageToMerge.Meta = Meta{
		Charset:      "charset_2",
		Title:        "title_2",
		Description:  "description_2",
		Keywords:     []string{"keyword_3", "keyword_4"},
		CanonicalURL: "http://canonicalurl_2",
		Icons: []Icon{{
			Rel:    "shorcut icon",
			Href:   "http://icon2",
			Type:   "image/jpeg",
			Color:  "#000",
			Width:  32,
			Height: 32,
		}},
		ImageSource: "http://imagesource_2",
		Robots:      "noindex,nofollow",
		ViewPort:    "viewport_2",
		HrefLang: []HrefLang{
			{
				Href:   "hreflang_href_3",
				Locale: "hreflang_locale_4",
			},
			{
				Href:   "hreflang_href_3",
				Locale: "hreflang_locale_4",
			},
		},
		Alternates: []Alternate{
			{
				Href:  "alternate_href_3",
				Type:  "alternate_type_3",
				Title: "alternate_title_3",
			},
			{
				Href:  "alternate_href_4",
				Type:  "alternate_type_4",
				Title: "alternate_title_4",
			},
		},
		Stylesheets: []Stylesheet{
			{
				Href:  "stylesheet_href_3",
				Media: "stylesheet_media_3",
			},
			{
				Href:  "stylesheet_href_4",
				Media: "stylesheet_media_4",
			},
		},
		OpenGraph: OpenGraph{
			Title:       "opengraph_title_2",
			Description: "opengraph_description_2",
			Image:       "http://opengraph_image_2",
			URL:         "http://opengraph_url_2",
			SiteName:    "opengraph_sitename_2",
		},
		Twitter: Twitter{
			Card:     "twitter_card_2",
			ImageAlt: "twitter_image_alt_2",
			Site:     "twitter_site_2",
		},
	}
	pageToMerge.Sitemap = weburl.Sitemap{
		Included:        false,
		Priority:        0.5,
		UpdateFrequency: weburl.SitemapUpdateFrequencyDaily,
	}
	pageToMerge.Key = "key_2"

	// Merge the page structs.
	page.Merge(*pageToMerge)

	// Check the fields.
	if page.Locale != pageToMerge.Locale {
		t.Error("To merge a Page struct must modified the property 'Locale'.")
	}
	if page.Timezone != pageToMerge.Timezone {
		t.Error("To merge a Page struct must modified the property 'Timezone'.")
	}
	if page.Title != pageToMerge.Title {
		t.Error("To merge a Page struct must modified the property 'Title'.")
	}
	if page.Template != pageToMerge.Template {
		t.Error("To merge a Page struct must modified the property 'Template'.")
	}
	if page.AccessRight != pageToMerge.AccessRight {
		t.Error("To merge a Page struct must modified the property 'AccessRight'.")
	}
	if page.Key != pageToMerge.Key {
		t.Error("To merge a Page struct must modified the property 'Key'.")
	}

	paths := map[string]string{}
	for locale, path := range page.Paths {
		paths[locale] = path
	}
	for locale, path := range pageToMerge.Paths {
		paths[locale] = path
	}

	if !reflect.DeepEqual(page.Paths, paths) {
		t.Error("To merge a Page struct must merge the 'Paths'.")
	}

	scriptFiles := append(originalScriptFiles, pageToMerge.ScriptFiles...)

	if !reflect.DeepEqual(page.ScriptFiles, scriptFiles) {
		t.Error("To merge a Page struct must merge the 'ScriptFiles'.")
	}

	if page.Meta.Title != pageToMerge.Meta.Title {
		t.Error("To merge a Page struct must modified the property 'Title'.")
	}

	if page.Sitemap.Priority != pageToMerge.Sitemap.Priority {
		t.Error("To merge a Page struct must modified the property 'Sitemap'.")
	}
}

func TestPageMergeEmptyFields(t *testing.T) {
	// Create a new page struct.
	page := generatePage()

	originalLocale := page.Locale
	originalTimezone := page.Timezone
	originalTitle := page.Title
	orignalTemplate := page.Template
	orignalAccessRight := page.AccessRight
	orignalPaths := page.Paths
	orignalScriptFiles := page.ScriptFiles
	originalMetaTitle := page.Meta.Title
	originalSitemapPriority := page.Sitemap.Priority
	originalKey := page.Key

	// Create another page struct.
	pageToMerge := NewPage()

	// Merge the page structs.
	page.Merge(*pageToMerge)

	// Check the fields.
	if page.Locale != originalLocale {
		t.Error("To merge a Page struct must not modified the property 'Locale' if it is not provided.")
	}
	if page.Timezone != originalTimezone {
		t.Error("To merge a Page struct must not modified the property 'Timezone' if it is not provided.")
	}
	if page.Title != originalTitle {
		t.Error("To merge a Page struct must not modified the property 'Title' if it is not provided.")
	}
	if page.Template != orignalTemplate {
		t.Error("To merge a Page struct must not modified the property 'Template' if it is not provided.")
	}
	if page.AccessRight != orignalAccessRight {
		t.Error("To merge a Page struct must not modified the property 'AccessRight' if it is not provided.")
	}

	if !reflect.DeepEqual(page.Paths, orignalPaths) {
		t.Error("To merge a Page struct must not modified the property 'Paths' if it is not provided.")
	}
	if !reflect.DeepEqual(page.ScriptFiles, orignalScriptFiles) {
		t.Error("To merge a Page struct must not modified the property 'ScriptFiles' if it is not provided.")
	}

	if page.Meta.Title != originalMetaTitle {
		t.Error("To merge a Page struct must not modified the property 'Meta' if it is not provided.")
	}
	if page.Sitemap.Priority != originalSitemapPriority {
		t.Error("To merge a Page struct must not modified the property 'Sitemap' if it is not provided.")
	}

	if page.Key == originalKey {
		t.Error("To merge a Page struct must modified the property 'Key'.")
	}
}

func generatePage() *Page {
	page := NewPage()
	page.Locale = "fr-fr"
	page.Timezone = "Africa/Lagos"
	page.Title = "title"
	page.Paths = map[string]string{
		"fr-fr": "/path-fr",
		"en-us": "/path-en",
	}
	page.Template = PageTemplate{
		Decorator: "template_decorator",
		Main:      "template_main",
	}
	page.AccessRight = AccessRightSignedIn
	page.ScriptFiles = []Script{
		{
			Src: "script_1.js",
		},
		{
			Src: "script_2.js",
		},
	}
	page.Meta = Meta{
		Charset:      "charset",
		Title:        "title",
		Description:  "description",
		Keywords:     []string{"keyword_1", "keyword_2"},
		CanonicalURL: "http://canonicalurl",
		Icons: []Icon{{
			Rel:    "icon",
			Href:   "http://icon",
			Type:   "image/png",
			Color:  "#fff",
			Width:  16,
			Height: 16,
		}},
		ImageSource: "http://imagesource",
		Robots:      "index,follow",
		ViewPort:    "viewport",
		HrefLang: []HrefLang{
			{
				Href:   "hreflang_href_1",
				Locale: "hreflang_locale_1",
			},
			{
				Href:   "hreflang_href_2",
				Locale: "hreflang_locale_2",
			},
		},
		Alternates: []Alternate{
			{
				Href:  "alternate_href_1",
				Type:  "alternate_type_1",
				Title: "alternate_title_1",
			},
			{
				Href:  "alternate_href_2",
				Type:  "alternate_type_2",
				Title: "alternate_title_2",
			},
		},
		Stylesheets: []Stylesheet{
			{
				Href:  "stylesheet_href_1",
				Media: "stylesheet_media_1",
			},
			{
				Href:  "stylesheet_href_2",
				Media: "stylesheet_media_2",
			},
		},
		OpenGraph: OpenGraph{
			Title:       "opengraph_title",
			Description: "opengraph_description",
			Image:       "http://opengraph_image",
			URL:         "http://opengraph_url",
			SiteName:    "opengraph_sitename",
		},
		Twitter: Twitter{
			Card:     "twitter_card",
			ImageAlt: "twitter_image_alt",
			Site:     "twitter_site",
		},
	}
	page.Sitemap = weburl.Sitemap{
		Included:        true,
		Priority:        1.0,
		UpdateFrequency: weburl.SitemapUpdateFrequencyHourly,
	}
	page.Partial = true
	page.Key = "key"

	return page
}

func TestNewPage(t *testing.T) {
	page := NewPage()

	if page.Vars == nil || page.Content.Vars == nil || page.Partial {
		t.Errorf("NewPage() = %+v, want initialized variables and a full page", page)
	}
}

func TestPageUpdateMetaRobots(t *testing.T) {
	tests := []struct {
		name     string
		robots   string
		included bool
		want     string
	}{
		{"included, empty", "", true, "index"},
		{"included, noindex", "noindex,follow", true, "index,follow"},
		{"included, no index directive", "follow", true, "follow,index"},
		{"included, already index", "index,follow", true, "index,follow"},
		{"included, uppercase", "NoIndex", true, "index"},
		{"excluded, empty", "", false, "noindex"},
		{"excluded, index", "index,follow", false, "noindex,follow"},
		{"excluded, no index directive", "follow", false, "follow,noindex"},
		{"excluded, already noindex", "noindex,follow", false, "noindex,follow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{Meta: Meta{Robots: tt.robots}, Sitemap: weburl.Sitemap{Included: tt.included}}

			page.UpdateMetaRobots()

			if page.Meta.Robots != tt.want {
				t.Errorf("Robots = %q, want %q", page.Meta.Robots, tt.want)
			}
		})
	}
}

func TestPagePreparePaths(t *testing.T) {
	page := &Page{Paths: map[string]string{
		"en-us": "/{{.Locale}}/home",
		"fr-fr": "/{{.Locale}}/accueil",
		"de-de": "/{{.Locale}}/start",
	}}

	page.PreparePaths("https://web.test", []string{"en-us", "fr-fr"}, map[string]string{"en": "en-us", "fr": "fr-fr", "de": "de-de"})

	wantPaths := map[string]string{"en-us": "/en-us/home", "fr-fr": "/fr-fr/accueil"}
	if !reflect.DeepEqual(page.Paths, wantPaths) {
		t.Errorf("Paths = %v, want %v", page.Paths, wantPaths)
	}

	wantHrefLangs := []HrefLang{
		{Href: "https://web.test/en-us/home", Locale: "en-us"},
		{Href: "https://web.test/fr-fr/accueil", Locale: "fr-fr"},
		{Href: "https://web.test/en-us/home", Locale: "en"},
		{Href: "https://web.test/fr-fr/accueil", Locale: "fr"},
	}
	if !reflect.DeepEqual(page.Meta.HrefLang, wantHrefLangs) {
		t.Errorf("HrefLang = %v, want %v", page.Meta.HrefLang, wantHrefLangs)
	}
}

func TestPagePrepareWithLocale(t *testing.T) {
	setup(t)

	page := &Page{
		conf:  testConf,
		Title: "home.title",
		Meta: Meta{
			Title:       "default.title",
			Description: "unknown message",
			Keywords:    []string{"default.keyword", "plain"},
			ImageSource: "{{.StaticHost}}/image.png",
			Icons:       []Icon{{Href: "{{.WebHost}}/{{.Locale}}/icon.png"}},
			Alternates:  []Alternate{{Href: "{{.ApiHost}}/feed"}},
			Stylesheets: []Stylesheet{{Href: "{{.StaticHost}}/app.css"}},
			OpenGraph:   OpenGraph{Title: "default.title", SiteName: "site.name", Image: "{{.StaticHost}}/og.png", URL: "{{.WebHost}}"},
			Twitter:     Twitter{ImageAlt: "default.title"},
		},
		ScriptFiles: []Script{{Src: "{{.StaticHost}}/app.js"}},
	}

	page.PrepareWithLocale("fr-fr", "Europe/Paris", "https://web.test", "https://static.test", "https://api.test", []string{"en-us", "fr-fr", "de-de"})

	tests := []struct{ name, got, want string }{
		{"locale", page.Locale, "fr-fr"},
		{"timezone", page.Timezone, "Europe/Paris"},
		{"title", page.Title, "Accueil"},
		{"meta title", page.Meta.Title, "Défaut"},
		{"unknown message is kept", page.Meta.Description, "unknown message"},
		{"keyword", page.Meta.Keywords[0], "mot-clé"},
		{"plain keyword", page.Meta.Keywords[1], "plain"},
		{"image source", page.Meta.ImageSource, "https://static.test/image.png"},
		{"icon", page.Meta.Icons[0].Href, "https://web.test/fr-fr/icon.png"},
		{"alternate", page.Meta.Alternates[0].Href, "https://api.test/feed"},
		{"stylesheet", page.Meta.Stylesheets[0].Href, "https://static.test/app.css"},
		{"script", page.ScriptFiles[0].Src, "https://static.test/app.js"},
		{"open graph title", page.Meta.OpenGraph.Title, "Défaut"},
		{"open graph site name", page.Meta.OpenGraph.SiteName, "Site"},
		{"open graph image", page.Meta.OpenGraph.Image, "https://static.test/og.png"},
		{"open graph URL", page.Meta.OpenGraph.URL, "https://web.test"},
		{"twitter image alt", page.Meta.Twitter.ImageAlt, "Défaut"},
		{"open graph locale", page.Meta.OpenGraph.Locale, "fr_FR"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	if want := []string{"en_US", "de_DE"}; !reflect.DeepEqual(page.Meta.OpenGraph.AlternateLocales, want) {
		t.Errorf("AlternateLocales = %v, want %v", page.Meta.OpenGraph.AlternateLocales, want)
	}
}

func TestPageMerge_FillsEmptyMaps(t *testing.T) {
	page := &Page{}

	page.Merge(Page{
		Locale:      "fr-fr",
		Timezone:    "Europe/Paris",
		Paths:       map[string]string{"fr-fr": "/accueil"},
		Vars:        map[string]any{"a": 1},
		ScriptFiles: []Script{{Src: "a.js"}},
	})
	page.Merge(Page{
		Paths:       map[string]string{"en-us": "/home"},
		Vars:        map[string]any{"b": 2},
		ScriptFiles: []Script{{Src: "b.js"}},
	})

	if page.Locale != "fr-fr" || page.Timezone != "Europe/Paris" {
		t.Errorf("locale = %q, timezone = %q", page.Locale, page.Timezone)
	}
	if want := map[string]string{"fr-fr": "/accueil", "en-us": "/home"}; !reflect.DeepEqual(page.Paths, want) {
		t.Errorf("Paths = %v, want %v", page.Paths, want)
	}
	if want := map[string]any{"a": 1, "b": 2}; !reflect.DeepEqual(page.Vars, want) {
		t.Errorf("Vars = %v, want %v", page.Vars, want)
	}
	if want := []Script{{Src: "a.js"}, {Src: "b.js"}}; !reflect.DeepEqual(page.ScriptFiles, want) {
		t.Errorf("ScriptFiles = %v, want %v", page.ScriptFiles, want)
	}
}

func TestPageCopy_IsIndependent(t *testing.T) {
	page := generatePage()
	page.Vars = map[string]any{"a": 1}

	copied := page.Copy()
	copied.Paths["fr-fr"] = "changed"
	copied.Vars["a"] = 2
	copied.ScriptFiles[0].Src = "changed"
	copied.Meta.Keywords[0] = "changed"

	if page.Paths["fr-fr"] != "/path-fr" || page.Vars["a"] != 1 || page.ScriptFiles[0].Src != "script_1.js" || page.Meta.Keywords[0] != "keyword_1" {
		t.Errorf("modifying the copy modified the page: %+v", page)
	}
}
