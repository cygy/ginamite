// Package pagetest initializes the pages of the web server for the tests of the packages using them: the routes, the
// translations, the templates and the configuration.
package pagetest

import (
	"testing"

	"github.com/cygy/ginamite/common/config"
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

// Setup initializes the pages with the test routes, translations, templates and configuration, and restores the
// global state at the end of the test. It returns the Gin engine holding the templates.
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
	webtest.SetTemplates(t, map[string]string{
		"decorator.html": `<html>{{.IncludeMainTemplate}}</html>`,
		"main.html":      `{{.Content.Title}}|{{.Content.Error.Message}}|{{.Content.Info.Message}}|{{.Var "ID"}}`,
	}, engine)

	old := html.Main
	t.Cleanup(func() { html.Main = old })

	file := webtest.WriteFile(t, t.TempDir(), "routes.yaml", routes)
	html.Initialize(file, WebHost, StaticHost, "https://api.test", "UTC", Locales, map[string]string{"en": "en-us", "fr": "fr-fr"},
		CookieName, Secret, html.FacebookConfiguration{}, html.GoogleConfiguration{}, html.SocialNetworksConfiguration{}, false, true)

	return engine
}
