// Package html describes the pages of the website from the routes file, and renders them.
//
// The routes file defines the default page, the routes and the passthrough asynchronous calls to the API. Initialize
// loads it and prepares, for each route and each locale, a page cloned by Configuration.PageByKey for each request.
package html

import (
	"os"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/web/context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

// Configuration is the configuration of the routes, loaded from the routes file.
type Configuration struct {
	Default           *Page
	Routes            map[string]*Page
	Async             map[string]*Async
	urls              map[string]map[string]string // The index is the key of the route, then the locale.
	preprocessedPages map[string]map[string]*Page  // The index is the key of the route, then the locale.
}

// settings are the properties of the application used to prepare the pages.
type settings struct {
	webHost         string
	staticHost      string
	apiHost         string
	timezone        string
	locales         []string
	fallbackLocales map[string]string
	configuration   PageConfiguration
	betaVersion     bool
	production      bool
}

// Main is the configuration of the routes, set by Initialize.
var Main *Configuration

var (
	jwtCookieName string
	jwtSecret     string
)

// Initialize loads the routes file into Main and prepares the pages for each supported locale.
// The JWT cookie name and secret are used to find the user of a request.
// It panics, after logging the error, if the file can not be read or parsed.
func Initialize(file, webHost, staticHost, apiHost, applicationTimezone string, applicationLocales []string, fallbackLocales map[string]string, JWTCookieName, JWTSecret string, facebookConfiguration FacebookConfiguration, googleConfiguration GoogleConfiguration, socialNetworksConfiguration SocialNetworksConfiguration, betaVersion, isProduction bool) {
	Main = load(file)
	Main.prepare(settings{
		webHost:         webHost,
		staticHost:      staticHost,
		apiHost:         apiHost,
		timezone:        applicationTimezone,
		locales:         applicationLocales,
		fallbackLocales: fallbackLocales,
		configuration: PageConfiguration{
			Facebook:       facebookConfiguration,
			Google:         googleConfiguration,
			SocialNetworks: socialNetworksConfiguration,
		},
		betaVersion: betaVersion,
		production:  isProduction,
	})

	jwtCookieName = JWTCookieName
	jwtSecret = JWTSecret
}

// load reads the routes file. It panics, after logging the error, if the file can not be read or parsed.
func load(file string) *Configuration {
	source, err := os.ReadFile(file)
	if err != nil {
		log.WithFields(logrus.Fields{
			"path":  file,
			"error": err.Error(),
		}).Panic("unable to load routes")
	}

	conf := &Configuration{}
	if err := yaml.Unmarshal(source, conf); err != nil {
		log.WithFields(logrus.Fields{
			"path":  file,
			"error": err.Error(),
		}).Panic("unable to load routes")
	}

	return conf
}

// prepare computes the URLs and the localized pages of the routes, and the paths of the async calls.
// A route without definition inherits everything from the default page.
func (conf *Configuration) prepare(s settings) {
	if conf.Default == nil {
		conf.Default = NewPage()
	}

	conf.urls = map[string]map[string]string{}
	conf.preprocessedPages = map[string]map[string]*Page{}

	for key, page := range conf.Routes {
		if page == nil {
			page = NewPage()
		}

		prepared := conf.Default.Copy()
		prepared.Merge(*page)
		prepared.UpdateMetaRobots()
		prepared.PreparePaths(s.webHost, s.locales, s.fallbackLocales)
		prepared.Configuration = s.configuration
		prepared.Key = key
		prepared.BetaVersion = s.betaVersion
		prepared.Production = s.production

		conf.preprocessedPages[key] = make(map[string]*Page, len(s.locales))
		for _, locale := range s.locales {
			localized := prepared.Copy()
			localized.PrepareWithLocale(locale, s.timezone, s.webHost, s.staticHost, s.apiHost, s.locales)
			conf.preprocessedPages[key][locale] = localized
		}

		conf.urls[key] = map[string]string{}
		for _, hrefLang := range prepared.Meta.HrefLang {
			conf.urls[key][hrefLang.Locale] = hrefLang.Href
		}
	}

	for _, async := range conf.Async {
		async.Prepare(s.webHost, s.locales)
	}
}

// PageByKey returns a new page for the route key and the locale of the request,
// or nil if the route or the locale is unknown.
func (conf *Configuration) PageByKey(c *gin.Context, key string) *Page {
	localizedPage, ok := conf.preprocessedPages[key][context.GetLocale(c)]
	if !ok {
		return nil
	}

	page := localizedPage.Copy()
	page.InitializeData(c)

	return page
}

// Page returns a new page for the route targeted by the request, see context.SavePageKey.
func (conf *Configuration) Page(c *gin.Context) *Page {
	return conf.PageByKey(c, context.GetPageKey(c))
}

// IsPartial reports whether the page of the route key is a partial page for the locale.
// An unknown page is a partial page.
func (conf *Configuration) IsPartial(key, locale string) bool {
	if localizedPage, ok := conf.preprocessedPages[key][locale]; ok {
		return localizedPage.Partial
	}

	return true
}

// URL returns the absolute URL of the route key for the locale, or an empty string if there is none.
func (conf *Configuration) URL(key, locale string) string {
	return conf.urls[key][locale]
}

// AsyncURL returns the absolute URL of the async call key for the locale, or an empty string if there is none.
func (conf *Configuration) AsyncURL(key, locale string) string {
	if async, ok := conf.Async[key]; ok {
		return async.Paths[locale]
	}

	return ""
}
