package html

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

const routesFile = `
default:
  template:
    decorator: decorator.html
    main: main.html
  meta:
    charset: utf-8
    title: default.title
    description: default.description
    keywords:
      - default.keyword
    robots: follow
    image_source: "{{.StaticHost}}/image.png"
    opengraph:
      site_name: site.name
      image: "{{.StaticHost}}/og.png"
    twitter:
      card: summary
  scripts:
    - src: "{{.StaticHost}}/app.js"
routes:
  home:
    paths:
      en-us: "/{{.Locale}}/home"
      fr-fr: "/{{.Locale}}/accueil"
    title: home.title
    sitemap:
      included: true
      priority: 1
      update_frequency: daily
  account:
    paths:
      en-us: "/{{.Locale}}/account/:id"
    partial: true
    access: signedin
  empty:
async:
  like:
    paths:
      en-us: "/{{.Locale}}/like"
      de-de: "/{{.Locale}}/like"
    method: POST
    target: /likes/:id
`

const (
	testWebHost    = "https://web.test"
	testStaticHost = "https://static.test"
	testAPIHost    = "https://api.test/v1"
	testSecret     = "secret"
	testCookieName = "auth"
)

var (
	testLocales  = []string{"en-us", "fr-fr"}
	testFallback = map[string]string{"en": "en-us", "fr": "fr-fr"}
	testFormats  = map[string]config.LocaleConfiguration{
		"en-us": {DateFormat: "2006-01-02", TimeFormat: "15:04", DateTimeFormat: "2006-01-02 15:04"},
		"fr-fr": {DateFormat: "02/01/2006", TimeFormat: "15h04", DateTimeFormat: "02/01/2006 15h04"},
	}
)

// setup initializes the routes with the test routes file, the translations, the templates and the configuration,
// and restores the global state at the end of the test. It returns the Gin engine holding the templates.
func setup(t *testing.T) *gin.Engine {
	t.Helper()

	webtest.SetTranslations(t, map[string]map[string]string{
		"en-us": {"default.title": "Default", "default.description": "Default description", "default.keyword": "default keyword", "site.name": "Site", "home.title": "Home", "hello": "Hello {{.Name}}"},
		"fr-fr": {"default.title": "Défaut", "default.description": "Description", "default.keyword": "mot-clé", "site.name": "Site", "home.title": "Accueil", "hello": "Bonjour {{.Name}}"},
	})

	engine := gin.New()
	webtest.SetTemplates(t, map[string]string{
		"decorator.html": `<html>{{.IncludeMainTemplate}}</html>`,
		"main.html":      `<h1>{{.Content.Title}}</h1>{{.Localize "hello" (.Content.Vars.args)}}`,
	}, engine)

	oldMain, oldName, oldSecret := Main, jwtCookieName, jwtSecret
	t.Cleanup(func() { Main, jwtCookieName, jwtSecret = oldMain, oldName, oldSecret })

	file := webtest.WriteFile(t, t.TempDir(), "routes.yaml", routesFile)
	Initialize(file, testWebHost, testStaticHost, testAPIHost, "Europe/Paris", testLocales, testFallback, testCookieName, testSecret,
		FacebookConfiguration{Enabled: true, AppID: "fb"}, GoogleConfiguration{ClientID: "google"}, SocialNetworksConfiguration{Twitter: "@site"},
		testFormats, true, false)

	return engine
}

// requestFor returns a Gin context for a GET request on the locale, with the cookies.
func requestFor(locale string, cookies ...*http.Cookie) *gin.Context {
	c, _ := webtest.Context(http.MethodGet, "/", nil)
	c.Set("_locale", locale)
	for _, cookie := range cookies {
		c.Request.AddCookie(cookie)
	}

	return c
}
