// Package pagetest initializes the pages of the web server for the tests of the packages using them: the routes, the
// translations, the templates and the configuration.
package pagetest

import (
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/ctxlocalization"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

// Test settings.
const (
	WebHost    = "https://web.test"
	StaticHost = "https://static.test"
	CookieName = "auth"
	Secret     = "secret"
)

// Pages are the pages prepared by Setup, nil outside of a test using it.
var Pages *html.Configuration

// Locales are the locales supported by the test pages.
var Locales = []string{"en-us", "fr-fr"}

const routes = `
default:
  template:
    decorator: decorator.html
    main: main.html
routes:
  index:
    paths:
      en-us: "/{{.Locale}}/home"
      fr-fr: "/{{.Locale}}/accueil"
    title: page.title
  account.settings.terms:
    paths:
      en-us: "/{{.Locale}}/account/terms"
      fr-fr: "/{{.Locale}}/compte/conditions"
  error.internal_server:
    paths:
      en-us: "/{{.Locale}}/error/internal-server"
      fr-fr: "/{{.Locale}}/erreur/serveur"
  error.not_found:
    paths:
      en-us: "/{{.Locale}}/error/not-found"
      fr-fr: "/{{.Locale}}/erreur/introuvable"
  error.unauthorized:
    paths:
      en-us: "/{{.Locale}}/error/unauthorized"
      fr-fr: "/{{.Locale}}/erreur/non-autorise"
  error.authentication.invalid:
    paths:
      en-us: "/{{.Locale}}/error/invalid-authentication"
      fr-fr: "/{{.Locale}}/erreur/authentification-invalide"
  error.authentication.expired:
    paths:
      en-us: "/{{.Locale}}/error/expired-authentication"
      fr-fr: "/{{.Locale}}/erreur/authentification-expiree"
  error.authentication.revoked:
    paths:
      en-us: "/{{.Locale}}/error/revoked-authentication"
      fr-fr: "/{{.Locale}}/erreur/authentification-revoquee"
  partial:
    partial: true
    paths:
      en-us: "/{{.Locale}}/partial"
      fr-fr: "/{{.Locale}}/partiel"
  page:
    paths:
      en-us: "/{{.Locale}}/page/:id"
      fr-fr: "/{{.Locale}}/page/:id"
    title: page.title
async:
  like:
    paths:
      en-us: "/{{.Locale}}/async/like/:id"
      fr-fr: "/{{.Locale}}/async/aimer/:id"
    method: POST
    target: /likes/:id
  list:
    paths:
      en-us: "/{{.Locale}}/async/list"
    method: GET
    target: /list
`

// Setup initializes the pages with the test routes, translations, templates and configuration, see Pages, and restores
// the test state at the end of the test. The contexts created by webtest.Context hold the pages. It returns the Gin engine holding the templates, which saves the configuration
// in the context of the requests.
//
// A page is rendered as "<html>TITLE|ERROR|INFO|ID</html>", where ID is the variable "ID".
func Setup(t *testing.T) *gin.Engine {
	t.Helper()

	webtest.SetConfig(t, &config.Configuration{
		Cookie: config.CookieConfiguration{Name: CookieName, Domain: "web.test", Path: "/", TTL: 3600, Secure: true, HTTPOnly: true},
	})
	webtest.SetTranslations(t, map[string]map[string]string{
		"en-us": {
			"page.title":                                "Title",
			"tpl.error.internal_server.id":              "Error {{.ErrorID}}",
			"error.unauthorized.message":                "Not allowed",
			"error.unauthorized.invalid_token.reason":   "Invalid token",
			"error.unauthorized.invalid_token.recovery": "Sign in",
		},
		"fr-fr": {
			"page.title":                                "Titre",
			"tpl.error.internal_server.id":              "Erreur {{.ErrorID}}",
			"error.unauthorized.message":                "Interdit",
			"error.unauthorized.invalid_token.reason":   "Jeton invalide",
			"error.unauthorized.invalid_token.recovery": "Connectez-vous",
		},
	})

	engine := gin.New()
	engine.Use(ctxconfig.Inject(webtest.Config))
	templates := webtest.SetTemplates(t, map[string]string{
		"decorator.html": `<html>{{.IncludeMainTemplate}}</html>`,
		"main.html":      `{{.Content.Title}}|{{.Content.Error.Message}}|{{.Content.Info.Message}}|{{.Var "ID"}}`,
	}, engine)

	file := webtest.WriteFile(t, t.TempDir(), "routes.yaml", routes)
	pages, err := html.Load(file, html.Settings{
		WebHost:         WebHost,
		StaticHost:      StaticHost,
		APIHost:         "https://api.test",
		Timezone:        "UTC",
		Locales:         Locales,
		FallbackLocales: map[string]string{"en": "en-us", "fr": "fr-fr"},
		Production:      true,
		Localization:    webtest.Localization,
		Templates:       templates,
	})
	if err != nil {
		t.Fatal(err)
	}
	Pages = pages
	t.Cleanup(func() { Pages = nil })

	engine.Use(html.Inject(pages), ctxlocalization.Inject(webtest.Localization))
	webtest.OnContext(t, func(c *gin.Context) { html.Set(c, pages) })

	return engine
}
