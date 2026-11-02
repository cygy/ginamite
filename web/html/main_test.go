package html

import (
	"html/template"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/ctxconfig"

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

// testConf is the configuration of the routes prepared by setup.
var testConf *Configuration

// loadTestRoutes returns the configuration of the routes of the test file, with the given templates.
func loadTestRoutes(t *testing.T, file string, templates *template.Template) *Configuration {
	t.Helper()

	conf, err := Load(file, Settings{
		WebHost:         testWebHost,
		StaticHost:      testStaticHost,
		APIHost:         testAPIHost,
		Timezone:        "Europe/Paris",
		Locales:         testLocales,
		FallbackLocales: testFallback,
		LocaleFormats:   testFormats,
		Facebook:        FacebookConfiguration{Enabled: true, AppID: "fb"},
		Google:          GoogleConfiguration{ClientID: "google"},
		SocialNetworks:  SocialNetworksConfiguration{Twitter: "@site"},
		BetaVersion:     true,
		Localization:    webtest.Localization,
		Templates:       templates,
	})
	if err != nil {
		t.Fatal(err)
	}

	return conf
}

// setup prepares the configuration of the routes, see testConf, with the test routes file, the translations, the
// templates and the configuration of the contexts. It returns the Gin engine holding the templates.
func setup(t *testing.T) *gin.Engine {
	t.Helper()

	webtest.SetConfig(t, &config.Configuration{
		Cookie: config.CookieConfiguration{Name: testCookieName},
		JWT:    config.JWTConfiguration{Secret: testSecret},
	})
	webtest.SetTranslations(t, map[string]map[string]string{
		"en-us": {"default.title": "Default", "default.description": "Default description", "default.keyword": "default keyword", "site.name": "Site", "home.title": "Home", "hello": "Hello {{.Name}}"},
		"fr-fr": {"default.title": "Défaut", "default.description": "Description", "default.keyword": "mot-clé", "site.name": "Site", "home.title": "Accueil", "hello": "Bonjour {{.Name}}"},
	})

	engine := gin.New()
	engine.Use(ctxconfig.Inject(webtest.Config))
	templates := webtest.SetTemplates(t, map[string]string{
		"decorator.html": `<html>{{.IncludeMainTemplate}}</html>`,
		"main.html":      `<h1>{{.Content.Title}}</h1>{{.Localize "hello" (.Content.Vars.args)}}`,
	}, engine)

	testConf = loadTestRoutes(t, webtest.WriteFile(t, t.TempDir(), "routes.yaml", routesFile), templates)
	t.Cleanup(func() { testConf = nil })
	webtest.OnContext(t, func(c *gin.Context) { Set(c, testConf) })

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
