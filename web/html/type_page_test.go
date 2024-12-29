package html

import (
	"reflect"
	"testing"
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
	pageToMerge.Sitemap = Sitemap{
		Included:        false,
		Priority:        0.5,
		UpdateFrequency: SitemapUpdateFrequencyDaily,
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

	if page.Key != originalKey {
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
	page.Sitemap = Sitemap{
		Included:        true,
		Priority:        1.0,
		UpdateFrequency: SitemapUpdateFrequencyHourly,
	}
	page.Partial = true
	page.Key = "key"

	return page
}
