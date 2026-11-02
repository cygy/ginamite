package html

import (
	"fmt"
	"sort"
	"strings"
)

// Meta is the metadata of a HTML page.
type Meta struct {
	Charset      string
	Title        string
	Description  string
	Keywords     []string
	CanonicalURL string `yaml:"-"` // Defined by the constructor
	Icons        []Icon
	ImageSource  string `yaml:"image_source"`
	Robots       string
	ViewPort     string
	HrefLang     []HrefLang `yaml:"-"` // Defined by the constructor
	Alternates   []Alternate
	Stylesheets  []Stylesheet
	OpenGraph    OpenGraph
	Twitter      Twitter
}

// Copy returns a deep copy of the metadata.
func (meta *Meta) Copy() (copied *Meta) {
	copied = &Meta{}
	copied.Charset = meta.Charset
	copied.Title = meta.Title
	copied.Description = meta.Description
	copied.CanonicalURL = meta.CanonicalURL
	copied.ImageSource = meta.ImageSource
	copied.Robots = meta.Robots
	copied.ViewPort = meta.ViewPort
	copied.OpenGraph = meta.OpenGraph
	copied.Twitter = meta.Twitter

	copied.Keywords = make([]string, len(meta.Keywords))
	copy(copied.Keywords, meta.Keywords)

	copied.Icons = make([]Icon, len(meta.Icons))
	copy(copied.Icons, meta.Icons)

	copied.HrefLang = make([]HrefLang, len(meta.HrefLang))
	copy(copied.HrefLang, meta.HrefLang)

	copied.Alternates = make([]Alternate, len(meta.Alternates))
	copy(copied.Alternates, meta.Alternates)

	copied.Stylesheets = make([]Stylesheet, len(meta.Stylesheets))
	copy(copied.Stylesheets, meta.Stylesheets)

	return
}

// Merge overrides the properties of the metadata with the properties defined in toMerge,
// and appends its keywords, icons, alternates and stylesheets.
func (meta *Meta) Merge(toMerge Meta) {
	if len(toMerge.Charset) > 0 {
		meta.Charset = toMerge.Charset
	}
	if len(toMerge.Title) > 0 {
		meta.Title = toMerge.Title
	}
	if len(toMerge.Description) > 0 {
		meta.Description = toMerge.Description
	}
	if len(toMerge.CanonicalURL) > 0 {
		meta.CanonicalURL = toMerge.CanonicalURL
	}
	if len(toMerge.ImageSource) > 0 {
		meta.ImageSource = toMerge.ImageSource
	}
	if len(toMerge.Robots) > 0 {
		meta.Robots = toMerge.Robots
	}
	if len(toMerge.ViewPort) > 0 {
		meta.ViewPort = toMerge.ViewPort
	}
	meta.OpenGraph.Merge(toMerge.OpenGraph)
	meta.Twitter.Merge(toMerge.Twitter)
	if toMerge.Keywords != nil {
		meta.Keywords = append(meta.Keywords, toMerge.Keywords...)
	}
	if toMerge.Icons != nil {
		meta.Icons = append(meta.Icons, toMerge.Icons...)
	}
	if toMerge.HrefLang != nil {
		meta.HrefLang = append(meta.HrefLang, toMerge.HrefLang...)
	}
	if toMerge.Alternates != nil {
		meta.Alternates = append(meta.Alternates, toMerge.Alternates...)
	}
	if toMerge.Stylesheets != nil {
		meta.Stylesheets = append(meta.Stylesheets, toMerge.Stylesheets...)
	}
}

// JoinedKeywords returns a string containing the keywords separated by a comma.
func (meta *Meta) JoinedKeywords() string {
	return strings.Join(meta.Keywords, ", ")
}

// setUpHrefLangs defines the alternate URLs of the page: one for each locale of paths, then one for each fallback
// language, whose URL is the one of its locale. The order of the URLs is stable: by locale, then by language.
func (meta *Meta) setUpHrefLangs(paths map[string]string, fallbackLocales map[string]string, host string) {
	meta.HrefLang = []HrefLang{}

	locales := make([]string, 0, len(paths))
	for locale := range paths {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	for _, locale := range locales {
		meta.HrefLang = append(meta.HrefLang, HrefLang{
			Href:   fmt.Sprintf("%s%s", host, paths[locale]),
			Locale: locale,
		})
	}

	languages := make([]string, 0, len(fallbackLocales))
	for language := range fallbackLocales {
		languages = append(languages, language)
	}
	sort.Strings(languages)

	for _, language := range languages {
		path, ok := paths[fallbackLocales[language]]
		if !ok {
			continue
		}

		meta.HrefLang = append(meta.HrefLang, HrefLang{
			Href:   fmt.Sprintf("%s%s", host, path),
			Locale: language,
		})
	}
}

// setUpCanonicalURL defines the canonical URL as the alternate URL of the locale, if there is one.
func (meta *Meta) setUpCanonicalURL(currentLocale string) {
	for _, hrefLang := range meta.HrefLang {
		if hrefLang.Locale == currentLocale {
			meta.CanonicalURL = hrefLang.Href
			break
		}
	}
}
