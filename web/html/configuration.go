// Package html describes the pages of the website from the routes file, and renders them.
//
// The routes file defines the default page, the routes and the passthrough asynchronous calls to the API. Load
// loads it and prepares, for each route and each locale, a page cloned by Configuration.PageByKey for each request.
package html

import (
	"html/template"
	"os"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
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
	settings          Settings
}

// Settings are the properties of the application used to prepare and render the pages.
type Settings struct {
	WebHost         string
	StaticHost      string
	APIHost         string
	Timezone        string // The timezone of the pages, unless the user has one.
	Locales         []string
	FallbackLocales map[string]string
	LocaleFormats   map[string]config.LocaleConfiguration // The date and time formats, by locale.
	Facebook        FacebookConfiguration
	Google          GoogleConfiguration
	SocialNetworks  SocialNetworksConfiguration
	BetaVersion     bool
	Production      bool

	// Localization translates the messages of the pages, and gives their raw strings.
	Localization *localization.Engine

	// Templates are the HTML templates of the pages.
	Templates *template.Template
}

// Load loads the routes file and prepares the pages for each supported locale.
// It panics, after logging the error, if the file can not be read or parsed.
func Load(file string, settings Settings) *Configuration {
	conf := read(file)
	conf.settings = settings
	conf.prepare()

	return conf
}

// read reads the routes file. It panics, after logging the error, if the file can not be read or parsed.
func read(file string) *Configuration {
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
func (conf *Configuration) prepare() {
	s := conf.settings

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
		prepared.PreparePaths(s.WebHost, s.Locales, s.FallbackLocales)
		prepared.Configuration = PageConfiguration{Facebook: s.Facebook, Google: s.Google, SocialNetworks: s.SocialNetworks}
		prepared.conf = conf
		prepared.Key = key
		prepared.BetaVersion = s.BetaVersion
		prepared.Production = s.Production

		conf.preprocessedPages[key] = make(map[string]*Page, len(s.Locales))
		for _, locale := range s.Locales {
			localized := prepared.Copy()
			localized.PrepareWithLocale(locale, s.Timezone, s.WebHost, s.StaticHost, s.APIHost, s.Locales)
			conf.preprocessedPages[key][locale] = localized
		}

		conf.urls[key] = map[string]string{}
		for _, hrefLang := range prepared.Meta.HrefLang {
			conf.urls[key][hrefLang.Locale] = hrefLang.Href
		}
	}

	for _, async := range conf.Async {
		async.Prepare(s.WebHost, s.Locales)
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
