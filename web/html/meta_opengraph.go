package html

import (
	"fmt"
	"reflect"
	"strings"
)

// OpenGraph is the Open Graph metadata of a page.
type OpenGraph struct {
	Title            string
	Description      string
	Image            string
	URL              string
	SiteName         string `yaml:"site_name"`
	Type             string
	Locale           string
	AlternateLocales []string
}

// Merge overrides the properties with the ones defined in toMerge.
func (openGraph *OpenGraph) Merge(toMerge OpenGraph) {
	if len(toMerge.Title) > 0 {
		openGraph.Title = toMerge.Title
	}
	if len(toMerge.Description) > 0 {
		openGraph.Description = toMerge.Description
	}
	if len(toMerge.Image) > 0 {
		openGraph.Image = toMerge.Image
	}
	if len(toMerge.URL) > 0 {
		openGraph.URL = toMerge.URL
	}
	if len(toMerge.SiteName) > 0 {
		openGraph.SiteName = toMerge.SiteName
	}
	if len(toMerge.Type) > 0 {
		openGraph.Type = toMerge.Type
	}
	if len(toMerge.Locale) > 0 {
		openGraph.Locale = toMerge.Locale
	}
	if len(toMerge.AlternateLocales) > 0 {
		openGraph.AlternateLocales = toMerge.AlternateLocales
	}
}

// Equals reports whether the two OpenGraph properties are equal.
func (openGraph *OpenGraph) Equals(toCompare *OpenGraph) bool {
	return openGraph.Title == toCompare.Title &&
		openGraph.Description == toCompare.Description &&
		openGraph.Image == toCompare.Image &&
		openGraph.URL == toCompare.URL &&
		openGraph.SiteName == toCompare.SiteName &&
		openGraph.Type == toCompare.Type &&
		openGraph.Locale == toCompare.Locale &&
		reflect.DeepEqual(openGraph.AlternateLocales, toCompare.AlternateLocales)
}

// FormatOpenGraphLocale returns a locale in the Open Graph format: "fr-FR" and "fr-fr" become "fr_FR".
// Anything else than a language and a region is returned as is.
func FormatOpenGraphLocale(locale string) string {
	parts := strings.Split(locale, "-")
	if len(parts) != 2 {
		return locale
	}

	return fmt.Sprintf("%s_%s", strings.ToLower(parts[0]), strings.ToUpper(parts[1]))
}
