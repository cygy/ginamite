package html

import (
	"fmt"
	"strings"
)

// OpenGraph : struct of OpenGraph Meta properties.
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

// Merge : merges two OpenGraph structs.
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

// Equals : returns true if the two structs are equal.
func (openGraph *OpenGraph) Equals(toCompare *OpenGraph) bool {
	if openGraph.Title != toCompare.Title {
		return false
	}
	if openGraph.Description != toCompare.Description {
		return false
	}
	if openGraph.Image != toCompare.Image {
		return false
	}
	if openGraph.URL != toCompare.URL {
		return false
	}
	if openGraph.SiteName != toCompare.SiteName {
		return false
	}
	if openGraph.Type != toCompare.Type {
		return false
	}
	if openGraph.Locale != toCompare.Locale {
		return false
	}

	if len(openGraph.AlternateLocales) != len(toCompare.AlternateLocales) {
		return false
	}

	return true
}

// FormatOpenGraphLocale : returns a locale in the open graph format.
func FormatOpenGraphLocale(locale string) string {
	parts := strings.Split(locale, "-")
	if len(parts) != 2 {
		return locale
	}

	return fmt.Sprintf("%s_%s", strings.ToLower(parts[0]), strings.ToUpper(parts[1]))
}
