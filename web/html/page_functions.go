package html

import (
	"errors"
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
)

// The methods below are the helpers of the templates.

// Localize returns the translation of the message ID key in the locale of the page, or key if the locale is unknown.
// See localization.TranslateFunc for the arguments.
func (page *Page) Localize(key string, args ...any) string {
	translate := localization.Translate(page.Locale)
	if translate == nil {
		return key
	}

	return translate(key, args...)
}

// LocalizeDate returns the date of t formatted for the locale and in the timezone of the page.
func (page *Page) LocalizeDate(t time.Time) string {
	return page.formatTime(t, config.Main.Locales[page.Locale].DateFormat)
}

// LocalizeTime returns the time of t formatted for the locale and in the timezone of the page.
func (page *Page) LocalizeTime(t time.Time) string {
	return page.formatTime(t, config.Main.Locales[page.Locale].TimeFormat)
}

// LocalizeDateTime returns the date and the time of t formatted for the locale and in the timezone of the page.
func (page *Page) LocalizeDateTime(t time.Time) string {
	return page.formatTime(t, config.Main.Locales[page.Locale].DateTimeFormat)
}

// formatTime formats t with the layout, in the timezone of the page. An unknown timezone leaves t unchanged.
func (page *Page) formatTime(t time.Time, layout string) string {
	t, _ = timeInTimezone(t, page.Timezone)

	return t.Format(layout)
}

// FormattedTime returns a duration of s seconds as "MM:SS", or "HH:MM:SS" from one hour.
func (page *Page) FormattedTime(s uint) string {
	hours, minutes, seconds := s/3600, s%3600/60, s%60

	if hours > 0 {
		return fmt.Sprintf("%.2d:%.2d:%.2d", hours, minutes, seconds)
	}

	return fmt.Sprintf("%.2d:%.2d", minutes, seconds)
}

// URL returns the URL of the route key in the locale of the page, whose parameters, prefixed by a colon,
// are replaced with the values of the first map of vars.
func (page *Page) URL(key string, vars ...map[string]string) string {
	return page.replaceURLVars(Main.URL(key, page.Locale), vars...)
}

// AsyncURL returns the URL of the async call key in the locale of the page, whose parameters, prefixed by a colon,
// are replaced with the values of the first map of vars.
func (page *Page) AsyncURL(key string, vars ...map[string]string) string {
	return page.replaceURLVars(Main.AsyncURL(key, page.Locale), vars...)
}

// StaticURL returns the URL of the static asset at path.
func (page *Page) StaticURL(path string) string {
	return fmt.Sprintf("%s/%s", config.Main.Hosts.Static, path)
}

// HTMLSafe returns text as HTML, which is not escaped by the templates.
func (page *Page) HTMLSafe(text string) template.HTML {
	return template.HTML(text) //nolint:gosec // The caller states that the text is safe.
}

// HTMLEscape returns text with the HTML special characters escaped.
func (page *Page) HTMLEscape(text string) string {
	return html.EscapeString(text)
}

// Nl2br replaces the line feeds of text with <br/>.
func (page *Page) Nl2br(text string) string {
	return strings.ReplaceAll(text, "\n", "<br/>")
}

// HTMLEscapeWithNl escapes the HTML special characters of text, then replaces its line feeds with <br/>.
func (page *Page) HTMLEscapeWithNl(text string) template.HTML {
	return page.HTMLSafe(page.Nl2br(page.HTMLEscape(text)))
}

// IncludeMainTemplate returns the result of the main template, to include to the decorator template.
func (page *Page) IncludeMainTemplate() template.HTML {
	return template.HTML(page.Content.parsedMainTemplate) //nolint:gosec // The template escapes its own variables.
}

// Minus returns a - b.
func (page *Page) Minus(a, b int) int {
	return a - b
}

// Plus returns a + b.
func (page *Page) Plus(a, b int) int {
	return a + b
}

// Mult returns a * b.
func (page *Page) Mult(a, b int) int {
	return a * b
}

// Div returns a / b, or an error, which is an error of the template, if b is zero.
func (page *Page) Div(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}

	return a / b, nil
}
