package html

import (
	"fmt"
	"strings"
)

var (
	// LocaleKeyword is the name of the variable replaced by the locale in the routes file: "{{.Locale}}".
	LocaleKeyword = "Locale"
	// WebHostKeyword is the name of the variable replaced by the host of the web server.
	WebHostKeyword = "WebHost"
	// StaticHostKeyword is the name of the variable replaced by the host of the static assets.
	StaticHostKeyword = "StaticHost"
	// APIHostKeyword is the name of the variable replaced by the host of the API.
	APIHostKeyword = "ApiHost"
)

// Page is a HTML page.
type Page struct {
	Locale        string `yaml:"-"` // Defined by the constructor
	Timezone      string `yaml:"-"` // Defined by the constructor
	Paths         map[string]string
	Template      PageTemplate
	AccessRight   AccessRight `yaml:"access"`
	ScriptFiles   []Script    `yaml:"scripts"`
	Meta          Meta
	Title         string
	Sitemap       Sitemap
	Configuration PageConfiguration `yaml:"-"`       // Defined by the constructor
	Content       PageContent       `yaml:"-"`       // Defined by the each script
	Partial       bool              `yaml:"partial"` // true if this page is a partial page (loaded by an ajax call), false if it is a full page.
	Key           string            // The key defining the route in the 'routes' file.
	BetaVersion   bool
	Production    bool
	Vars          map[string]any // Common variables to all the pages.
}

// NewPage returns an empty page, without nil variables.
func NewPage() *Page {
	page := &Page{}
	page.Content.Vars = map[string]any{}
	page.Vars = map[string]any{}
	page.Partial = false

	return page
}

// Copy returns a deep copy of the page, except for its content, which is empty.
func (page *Page) Copy() (copied *Page) {
	copied = &Page{}
	copied.Locale = page.Locale
	copied.Timezone = page.Timezone
	copied.Template = page.Template
	copied.AccessRight = page.AccessRight
	copied.Sitemap = page.Sitemap
	copied.Meta = *page.Meta.Copy()
	copied.Title = page.Title
	copied.Configuration = page.Configuration
	copied.Partial = page.Partial
	copied.Key = page.Key
	copied.BetaVersion = page.BetaVersion
	copied.Production = page.Production

	copied.Paths = make(map[string]string, len(page.Paths))
	for key, value := range page.Paths {
		copied.Paths[key] = value
	}

	copied.Vars = make(map[string]any, len(page.Vars))
	for key, value := range page.Vars {
		copied.Vars[key] = value
	}

	copied.ScriptFiles = make([]Script, len(page.ScriptFiles))
	copy(copied.ScriptFiles, page.ScriptFiles)

	return
}

// Merge overrides the properties of the page with the properties defined in toMerge, merges the paths and the
// variables, and appends the script files.
// The partial property and the key are always overridden.
func (page *Page) Merge(toMerge Page) {
	if len(toMerge.Locale) > 0 {
		page.Locale = toMerge.Locale
	}
	if len(toMerge.Timezone) > 0 {
		page.Timezone = toMerge.Timezone
	}
	if len(toMerge.AccessRight) > 0 {
		page.AccessRight = toMerge.AccessRight
	}
	if len(toMerge.Title) > 0 {
		page.Title = toMerge.Title
	}

	page.Template.Merge(toMerge.Template)
	page.Meta.Merge(toMerge.Meta)

	if toMerge.Sitemap.IsDefined() {
		page.Sitemap = toMerge.Sitemap
	}

	if len(toMerge.Paths) > 0 {
		if page.Paths == nil {
			page.Paths = make(map[string]string, len(toMerge.Paths))
		}
		for key, value := range toMerge.Paths {
			page.Paths[key] = value
		}
	}

	if len(toMerge.Vars) > 0 {
		if page.Vars == nil {
			page.Vars = make(map[string]any, len(toMerge.Vars))
		}
		for key, value := range toMerge.Vars {
			page.Vars[key] = value
		}
	}

	if toMerge.ScriptFiles != nil {
		if page.ScriptFiles == nil {
			page.ScriptFiles = []Script{}
		}
		page.ScriptFiles = append(page.ScriptFiles, toMerge.ScriptFiles...)
	}

	page.Partial = toMerge.Partial
	page.Key = toMerge.Key
}

// UpdateMetaRobots makes the meta "robots" consistent with the sitemap: "index" if the page is included in the
// sitemap, "noindex" if not.
func (page *Page) UpdateMetaRobots() {
	robots := strings.ToLower(page.Meta.Robots)

	if page.Sitemap.Included {
		if strings.Contains(robots, "noindex") {
			robots = strings.ReplaceAll(robots, "noindex", "index")
		} else if len(robots) == 0 {
			robots = "index"
		} else if !strings.Contains(robots, "index") {
			robots = fmt.Sprintf("%s,index", robots)
		}
	} else if !strings.Contains(robots, "noindex") {
		if strings.Contains(robots, "index") {
			robots = strings.ReplaceAll(robots, "index", "noindex")
		} else if len(robots) == 0 {
			robots = "noindex"
		} else {
			robots = fmt.Sprintf("%s,noindex", robots)
		}
	}

	page.Meta.Robots = robots
}

// PreparePaths keeps the paths of the supported locales, replaces their locale keyword and defines the alternate URLs
// of the page.
func (page *Page) PreparePaths(webHost string, applicationLocales []string, fallbackLocales map[string]string) {
	// Remove the unsupported locales from the paths map.
	paths := map[string]string{}
	localeKey := fmt.Sprintf("{{.%s}}", LocaleKeyword)
	for _, applicationLocale := range applicationLocales {
		if path, ok := page.Paths[applicationLocale]; ok {
			paths[applicationLocale] = strings.Replace(path, localeKey, applicationLocale, 1)
		}
	}
	page.Paths = paths

	// Sets up the href lang to the metadata.
	page.Meta.setUpHrefLangs(page.Paths, fallbackLocales, webHost)
}

// PrepareWithLocale localizes the page: it saves the locale and the timezone, translates the title and the metadata,
// and replaces the host and the locale keywords of the URLs.
func (page *Page) PrepareWithLocale(currentLocale, currentTimezone, webHost, staticHost, apiHost string, supportedLocales []string) {
	// Save the locale and the timezone.
	page.Locale = currentLocale
	page.Timezone = currentTimezone

	// Localize some strings.
	page.Title = page.rawString(page.Title)

	page.Meta.Title = page.rawString(page.Meta.Title)
	page.Meta.Description = page.rawString(page.Meta.Description)
	page.Meta.OpenGraph.Title = page.rawString(page.Meta.OpenGraph.Title)
	page.Meta.OpenGraph.Description = page.rawString(page.Meta.OpenGraph.Description)
	page.Meta.OpenGraph.SiteName = page.rawString(page.Meta.OpenGraph.SiteName)
	page.Meta.Twitter.ImageAlt = page.rawString(page.Meta.Twitter.ImageAlt)

	// Localize the meta keywords.
	for i, keyword := range page.Meta.Keywords {
		page.Meta.Keywords[i] = page.rawString(keyword)
	}

	// Replace the variables to the URLs.
	variables := map[string]string{
		WebHostKeyword:    webHost,
		StaticHostKeyword: staticHost,
		APIHostKeyword:    apiHost,
		LocaleKeyword:     currentLocale,
	}
	for name, value := range variables {
		key := fmt.Sprintf("{{.%s}}", name)
		page.Meta.ImageSource = strings.Replace(page.Meta.ImageSource, key, value, 1)
		page.Meta.OpenGraph.Image = strings.Replace(page.Meta.OpenGraph.Image, key, value, 1)
		page.Meta.OpenGraph.URL = strings.Replace(page.Meta.OpenGraph.URL, key, value, 1)

		for i, icon := range page.Meta.Icons {
			icon.Href = strings.Replace(icon.Href, key, value, 1)
			page.Meta.Icons[i] = icon
		}
		for i, alternate := range page.Meta.Alternates {
			alternate.Href = strings.Replace(alternate.Href, key, value, 1)
			page.Meta.Alternates[i] = alternate
		}
		for i, stylesheet := range page.Meta.Stylesheets {
			stylesheet.Href = strings.Replace(stylesheet.Href, key, value, 1)
			page.Meta.Stylesheets[i] = stylesheet
		}
		for i, scriptFile := range page.ScriptFiles {
			scriptFile.Src = strings.Replace(scriptFile.Src, key, value, 1)
			page.ScriptFiles[i] = scriptFile
		}
	}

	// Set the locales up in the open graph tags.
	page.Meta.OpenGraph.Locale = FormatOpenGraphLocale(currentLocale)
	page.Meta.OpenGraph.AlternateLocales = []string{}
	for _, locale := range supportedLocales {
		if locale != currentLocale {
			page.Meta.OpenGraph.AlternateLocales = append(page.Meta.OpenGraph.AlternateLocales, FormatOpenGraphLocale(locale))
		}
	}
}
