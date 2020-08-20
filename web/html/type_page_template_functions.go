package html

import (
	"fmt"
	"html"
	"html/template"
	"math"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
)

// Localize : returns a localized string (used into the templates).
func (page *Page) Localize(key string, args ...interface{}) string {
	return localization.Translate(page.Locale)(key, args...)
}

// LocalizeDate : returns a localized date (used into the templates).
func (page *Page) LocalizeDate(t time.Time) string {
	timeInTimezone, _ := timeInTimezone(t, page.Timezone)
	return timeInTimezone.Format(config.Main.Locales[page.Locale].DateFormat)
}

// LocalizeTime : returns a localized time (used into the templates).
func (page *Page) LocalizeTime(t time.Time) string {
	timeInTimezone, _ := timeInTimezone(t, page.Timezone)
	return timeInTimezone.Format(config.Main.Locales[page.Locale].TimeFormat)
}

// LocalizeDateTime : returns a localized date and time (used into the templates).
func (page *Page) LocalizeDateTime(t time.Time) string {
	timeInTimezone, _ := timeInTimezone(t, page.Timezone)
	return timeInTimezone.Format(config.Main.Locales[page.Locale].DateTimeFormat)
}

// FormattedTime : returns a formatted time (used into the templates).
func (page *Page) FormattedTime(s uint) string {
	seconds := s
	hours := uint(math.Floor(float64(seconds) / 3600))
	seconds = seconds % 3600
	minutes := uint(math.Floor(float64(seconds) / 60))
	seconds = seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%.2d:%.2d:%.2d", hours, minutes, seconds)
	}

	return fmt.Sprintf("%.2d:%.2d", minutes, seconds)
}

// URL : returns a localized url (used into the templates).
func (page *Page) URL(key string, vars ...map[string]string) string {
	url := Main.URL(key, page.Locale)
	return page.replaceURLVars(url, vars...)
}

// AsyncURL : returns a localized url (used into the templates).
func (page *Page) AsyncURL(key string, vars ...map[string]string) string {
	url := Main.AsyncURL(key, page.Locale)
	return page.replaceURLVars(url, vars...)
}

// StaticURL : returns the url of a static asset (used into the templates).
func (page *Page) StaticURL(path string) string {
	return fmt.Sprintf("%s/%s", config.Main.Hosts.Static, path)
}

// HTMLSafe : returns a HTML unescaped string (used into the templates).
func (page *Page) HTMLSafe(text string) template.HTML {
	return template.HTML(text)
}

// HTMLEscape : returns a HTML escaped string (used into the templates).
func (page *Page) HTMLEscape(text string) string {
	return html.EscapeString(text)
}

// Nl2br : replaces the \n to a <br>.
func (page *Page) Nl2br(text string) string {
	return strings.Replace(text, "\n", "<br/>", -1)
}

// HTMLEscapeWithNl : replaces the \n to a <br>.
func (page *Page) HTMLEscapeWithNl(text string) template.HTML {
	return page.HTMLSafe(page.Nl2br(page.HTMLEscape(text)))
}

// IncludeMainTemplate : returns the content of the page, used to include the parsed Main template to the Decorator template.
func (page *Page) IncludeMainTemplate() template.HTML {
	return template.HTML(page.Content.parsedMainTemplate)
}

// Minus : arithmetic function.
func (page *Page) Minus(a, b int) int {
	return a - b
}

// Plus : arithmetic function.
func (page *Page) Plus(a, b int) int {
	return a + b
}

// Mult : arithmetic function.
func (page *Page) Mult(a, b int) int {
	return a * b
}

// Div : arithmetic function.
func (page *Page) Div(a, b int) int {
	return a / b
}
