package html

import (
	"html/template"
	"testing"
	"time"
)

func TestPage_Localize(t *testing.T) {
	setup(t)

	tests := []struct {
		name   string
		locale string
		key    string
		args   []any
		want   string
	}{
		{"english", "en-us", "hello", []any{map[string]string{"Name": "Ann"}}, "Hello Ann"},
		{"french", "fr-fr", "hello", []any{map[string]string{"Name": "Ann"}}, "Bonjour Ann"},
		{"unknown message", "en-us", "missing", nil, "missing"},
		{"unknown locale", "de-de", "hello", nil, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{Locale: tt.locale, conf: testConf}

			if got := page.Localize(tt.key, tt.args...); got != tt.want {
				t.Errorf("Localize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPage_LocalizeDateTime(t *testing.T) {
	at := time.Date(2026, time.July, 1, 23, 30, 0, 0, time.UTC)
	tests := []struct {
		name     string
		locale   string
		timezone string
		date     string
		time     string
		dateTime string
	}{
		{"english in utc", "en-us", "UTC", "2026-07-01", "23:30", "2026-07-01 23:30"},
		{"french in Paris", "fr-fr", "Europe/Paris", "02/07/2026", "01h30", "02/07/2026 01h30"},
		{"unknown timezone", "en-us", "Mars/Olympus", "2026-07-01", "23:30", "2026-07-01 23:30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{Locale: tt.locale, Timezone: tt.timezone, conf: &Configuration{settings: Settings{LocaleFormats: testFormats}}}

			if got := page.LocalizeDate(at); got != tt.date {
				t.Errorf("LocalizeDate() = %q, want %q", got, tt.date)
			}
			if got := page.LocalizeTime(at); got != tt.time {
				t.Errorf("LocalizeTime() = %q, want %q", got, tt.time)
			}
			if got := page.LocalizeDateTime(at); got != tt.dateTime {
				t.Errorf("LocalizeDateTime() = %q, want %q", got, tt.dateTime)
			}
		})
	}
}

func TestPage_FormattedTime(t *testing.T) {
	tests := []struct {
		seconds uint
		want    string
	}{
		{0, "00:00"},
		{59, "00:59"},
		{61, "01:01"},
		{3599, "59:59"},
		{3600, "01:00:00"},
		{3661, "01:01:01"},
		{360000, "100:00:00"},
	}
	for _, tt := range tests {
		if got := (&Page{}).FormattedTime(tt.seconds); got != tt.want {
			t.Errorf("FormattedTime(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestPage_URLs(t *testing.T) {
	setup(t)

	page := &Page{Locale: "en-us", conf: testConf}

	tests := []struct{ name, got, want string }{
		{"URL", page.URL("home"), testWebHost + "/en-us/home"},
		{"URL with parameter", page.URL("account", map[string]string{"id": "42"}), testWebHost + "/en-us/account/42"},
		{"unknown URL", page.URL("unknown"), ""},
		{"async URL", page.AsyncURL("like"), testWebHost + "/en-us/like"},
		{"static URL", page.StaticURL("css/app.css"), testStaticHost + "/css/app.css"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestPage_HTMLHelpers(t *testing.T) {
	page := &Page{}

	tests := []struct{ name, got, want string }{
		{"HTMLSafe", string(page.HTMLSafe("<b>x</b>")), "<b>x</b>"},
		{"HTMLEscape", page.HTMLEscape(`<a href="x">&</a>`), "&lt;a href=&#34;x&#34;&gt;&amp;&lt;/a&gt;"},
		{"Nl2br", page.Nl2br("a\nb\n"), "a<br/>b<br/>"},
		{"HTMLEscapeWithNl", string(page.HTMLEscapeWithNl("<a>\nb")), "&lt;a&gt;<br/>b"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestPage_IncludeMainTemplate(t *testing.T) {
	page := &Page{Content: PageContent{parsedMainTemplate: "<p>main</p>"}}

	if got := page.IncludeMainTemplate(); got != template.HTML("<p>main</p>") {
		t.Errorf("IncludeMainTemplate() = %q", got)
	}
}

func TestPage_Arithmetic(t *testing.T) {
	page := &Page{}

	if got := page.Plus(2, 3); got != 5 {
		t.Errorf("Plus() = %d", got)
	}
	if got := page.Minus(2, 3); got != -1 {
		t.Errorf("Minus() = %d", got)
	}
	if got := page.Mult(2, 3); got != 6 {
		t.Errorf("Mult() = %d", got)
	}
	if got, err := page.Div(7, 2); err != nil || got != 3 {
		t.Errorf("Div() = %d, %v", got, err)
	}
	if _, err := page.Div(1, 0); err == nil {
		t.Error("Div() by zero must return an error")
	}
}
