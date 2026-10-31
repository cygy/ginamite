package html

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestInitialize(t *testing.T) {
	setup(t)

	if jwtCookieName != testCookieName || jwtSecret != testSecret {
		t.Errorf("the JWT settings are %q and %q", jwtCookieName, jwtSecret)
	}

	t.Run("pages", func(t *testing.T) {
		if got := len(Main.preprocessedPages); got != 3 {
			t.Fatalf("%d pages prepared, want 3", got)
		}

		home := Main.preprocessedPages["home"]["fr-fr"]
		if home.Locale != "fr-fr" || home.Timezone != "Europe/Paris" || home.Key != "home" {
			t.Errorf("locale = %q, timezone = %q, key = %q", home.Locale, home.Timezone, home.Key)
		}
		if home.Title != "Accueil" || home.Meta.Title != "Défaut" || home.Meta.Description != "Description" {
			t.Errorf("title = %q, meta title = %q, meta description = %q", home.Title, home.Meta.Title, home.Meta.Description)
		}
		if !reflect.DeepEqual(home.Meta.Keywords, []string{"mot-clé"}) {
			t.Errorf("keywords = %v", home.Meta.Keywords)
		}
		if home.Meta.ImageSource != testStaticHost+"/image.png" || home.ScriptFiles[0].Src != testStaticHost+"/app.js" {
			t.Errorf("image source = %q, script = %q", home.Meta.ImageSource, home.ScriptFiles[0].Src)
		}
		if home.Meta.OpenGraph.Locale != "fr_FR" || !reflect.DeepEqual(home.Meta.OpenGraph.AlternateLocales, []string{"en_US"}) {
			t.Errorf("open graph locale = %q, alternates = %v", home.Meta.OpenGraph.Locale, home.Meta.OpenGraph.AlternateLocales)
		}
		if !home.BetaVersion || home.Production {
			t.Errorf("beta = %v, production = %v", home.BetaVersion, home.Production)
		}
		if !home.Configuration.Facebook.Enabled || home.Configuration.Google.ClientID != "google" || home.Configuration.SocialNetworks.Twitter != "@site" {
			t.Errorf("configuration = %+v", home.Configuration)
		}
		if home.Template.Decorator != "decorator.html" || home.Template.Main != "main.html" {
			t.Errorf("templates = %+v", home.Template)
		}
	})

	t.Run("robots", func(t *testing.T) {
		tests := map[string]string{
			"home":    "follow,index",
			"account": "follow,noindex",
			"empty":   "follow,noindex",
		}
		for key, want := range tests {
			if got := Main.preprocessedPages[key]["en-us"].Meta.Robots; got != want {
				t.Errorf("robots of %q = %q, want %q", key, got, want)
			}
		}
	})

	t.Run("urls", func(t *testing.T) {
		tests := []struct{ key, locale, want string }{
			{"home", "en-us", testWebHost + "/en-us/home"},
			{"home", "fr-fr", testWebHost + "/fr-fr/accueil"},
			{"home", "fr", testWebHost + "/fr-fr/accueil"},
			{"home", "en", testWebHost + "/en-us/home"},
			{"account", "en-us", testWebHost + "/en-us/account/:id"},
			{"account", "fr-fr", ""},
			{"unknown", "en-us", ""},
		}
		for _, tt := range tests {
			if got := Main.URL(tt.key, tt.locale); got != tt.want {
				t.Errorf("URL(%q, %q) = %q, want %q", tt.key, tt.locale, got, tt.want)
			}
		}
	})

	t.Run("async", func(t *testing.T) {
		if got := Main.AsyncURL("like", "en-us"); got != testWebHost+"/en-us/like" {
			t.Errorf("AsyncURL() = %q", got)
		}
		if got := Main.AsyncURL("like", "de-de"); got != "" {
			t.Errorf("AsyncURL() of an unsupported locale = %q, want empty", got)
		}
		if got := Main.AsyncURL("unknown", "en-us"); got != "" {
			t.Errorf("AsyncURL() of an unknown call = %q, want empty", got)
		}
	})
}

func TestInitialize_Panics(t *testing.T) {
	tests := []struct {
		name string
		file func(t *testing.T) string
	}{
		{"missing file", func(t *testing.T) string { return t.TempDir() + "/missing.yaml" }},
		{"invalid file", func(t *testing.T) string { return webtest.WriteFile(t, t.TempDir(), "routes.yaml", "routes: [") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webtest.Hook.Reset()
			file := tt.file(t)

			defer func() {
				if recover() == nil {
					t.Error("Initialize() must panic")
				}
				if entry := webtest.Hook.LastEntry(); entry == nil || entry.Message != "unable to load routes" || entry.Data["path"] != file {
					t.Errorf("logged entry = %+v", entry)
				}
			}()

			Initialize(file, "", "", "", "UTC", testLocales, testFallback, "", "", FacebookConfiguration{}, GoogleConfiguration{}, SocialNetworksConfiguration{}, nil, false, false)
		})
	}
}

func TestInitialize_WithoutDefaultPage(t *testing.T) {
	old := Main
	t.Cleanup(func() { Main = old })
	webtest.SetTranslations(t, map[string]map[string]string{"en-us": {"a": "b"}, "fr-fr": {"a": "b"}})

	file := webtest.WriteFile(t, t.TempDir(), "routes.yaml", "routes:\n  home:\n    paths:\n      en-us: /home\n")
	Initialize(file, testWebHost, "", "", "UTC", testLocales, testFallback, "", "", FacebookConfiguration{}, GoogleConfiguration{}, SocialNetworksConfiguration{}, nil, false, true)

	page := Main.preprocessedPages["home"]["en-us"]
	if page == nil || !page.Production || Main.URL("home", "en-us") != testWebHost+"/home" {
		t.Errorf("page = %+v, url = %q", page, Main.URL("home", "en-us"))
	}
}

func TestConfiguration_PageByKey(t *testing.T) {
	setup(t)

	t.Run("unknown", func(t *testing.T) {
		tests := []struct{ key, locale string }{
			{"unknown", "en-us"},
			{"home", "de-de"},
		}
		for _, tt := range tests {
			if page := Main.PageByKey(requestFor(tt.locale), tt.key); page != nil {
				t.Errorf("PageByKey(%q, %q) = %+v, want nil", tt.key, tt.locale, page)
			}
		}
	})

	t.Run("copy", func(t *testing.T) {
		first := Main.PageByKey(requestFor("en-us"), "home")
		first.Title = "changed"
		first.Meta.Keywords[0] = "changed"
		first.SetVar("a", 1)

		second := Main.PageByKey(requestFor("en-us"), "home")
		if second.Title != "Home" || second.Meta.Keywords[0] != "default keyword" || second.Var("a") != nil {
			t.Errorf("the prepared page is shared: %+v", second)
		}
		if second.Content.Title != "Home" {
			t.Errorf("Content.Title = %q, want Home", second.Content.Title)
		}
	})

	t.Run("current user", func(t *testing.T) {
		token := webtest.SignToken(t, testSecret, map[string]any{"sub": "u1"})
		c := requestFor("en-us", &http.Cookie{Name: testCookieName, Value: token})

		if page := Main.PageByKey(c, "home"); page.Content.CurrentUser.ID != "u1" {
			t.Errorf("CurrentUser = %+v", page.Content.CurrentUser)
		}
	})
}

func TestConfiguration_Page(t *testing.T) {
	setup(t)

	c := requestFor("fr-fr")
	if page := Main.Page(c); page != nil {
		t.Errorf("Page() without page key = %+v, want nil", page)
	}

	c.Set("_pagekey", "home")
	if page := Main.Page(c); page == nil || page.Key != "home" || page.Locale != "fr-fr" {
		t.Errorf("Page() = %+v", page)
	}
}

func TestConfiguration_IsPartial(t *testing.T) {
	setup(t)

	tests := []struct {
		key, locale string
		want        bool
	}{
		{"home", "en-us", false},
		{"account", "en-us", true},
		{"unknown", "en-us", true},
		{"home", "de-de", true},
	}
	for _, tt := range tests {
		if got := Main.IsPartial(tt.key, tt.locale); got != tt.want {
			t.Errorf("IsPartial(%q, %q) = %v, want %v", tt.key, tt.locale, got, tt.want)
		}
	}
}

func TestConfiguration_PageByKey_SharesNothingBetweenLocales(t *testing.T) {
	setup(t)

	en := Main.PageByKey(requestFor("en-us"), "home")
	fr := Main.PageByKey(requestFor("fr-fr"), "home")

	if en.Title == fr.Title || strings.Contains(en.Meta.OpenGraph.Locale, fr.Meta.OpenGraph.Locale) {
		t.Errorf("en = %q/%q, fr = %q/%q", en.Title, en.Meta.OpenGraph.Locale, fr.Title, fr.Meta.OpenGraph.Locale)
	}
}
