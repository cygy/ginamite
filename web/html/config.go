package html

import (
	"os"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/web/context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v3"
)

// Configuration : struct of the configuration of the routes.
type Configuration struct {
	Default           *Page
	Routes            map[string]*Page
	Async             map[string]*Async
	urls              map[string]map[string]string `yaml:"-"`
	preprocessedPages map[string]map[string]*Page  `yaml:"-"` // i.e. preprocessedRoutes[index][fr-fr] returns a standard *Page struct
}

// Main : configuration of the routes: singleton.
var Main *Configuration

var (
	jwtCookieName string
	jwtSecret     string
)

// Initialize : initialize the configuration of the routes.
func Initialize(file, webHost, staticHost, apiHost, applicationTimezone string, applicationLocales []string, fallbackLocales map[string]string, JWTCookieName, JWTSecret string, facebookConfiguration FacebookConfiguration, googleConfiguration GoogleConfiguration, socialNetworksConfiguration SocialNetworksConfiguration, betaVersion, isProduction bool) {
	// Read the configuration file.
	source, err := os.ReadFile(file)
	if err != nil {
		log.WithFields(logrus.Fields{
			"path":  file,
			"error": err.Error(),
		}).Panic("unable to load routes")
	}

	// Create the main configuration object from the file.
	Main = &Configuration{}
	err = yaml.Unmarshal(source, Main)
	if err != nil {
		log.WithFields(logrus.Fields{
			"path":  file,
			"error": err.Error(),
		}).Panic("unable to load routes")
	}

	// Prepare all the URLs and the preprocessed routes.
	Main.urls = map[string]map[string]string{}
	Main.preprocessedPages = map[string]map[string]*Page{}

	// Prepare all the routes.
	for key, page := range Main.Routes {
		preparedPage := Main.Default.Copy()
		preparedPage.Merge(*page)
		preparedPage.UpdateMetaRobots()
		preparedPage.PreparePaths(webHost, applicationLocales, fallbackLocales)
		preparedPage.Configuration.Facebook = facebookConfiguration
		preparedPage.Configuration.Google = googleConfiguration
		preparedPage.Configuration.SocialNetworks = socialNetworksConfiguration
		preparedPage.Key = key
		preparedPage.BetaVersion = betaVersion
		preparedPage.Production = isProduction

		Main.preprocessedPages[key] = make(map[string]*Page, len(applicationLocales))
		for _, applicationLocale := range applicationLocales {
			localizedPage := preparedPage.Copy()
			localizedPage.PrepareWithLocale(applicationLocale, applicationTimezone, webHost, staticHost, apiHost, applicationLocales)
			Main.preprocessedPages[key][applicationLocale] = localizedPage
		}

		Main.urls[key] = map[string]string{}
		for _, hrefLang := range preparedPage.Meta.HrefLang {
			Main.urls[key][hrefLang.Locale] = hrefLang.Href
		}
	}

	// Prepare all the async calls.
	for _, async := range Main.Async {
		async.Prepare(webHost, applicationLocales)
	}

	// Save the JWT configuration.
	jwtCookieName = JWTCookieName
	jwtSecret = JWTSecret
}

// PageByKey : returns the standard page struct with the key.
func (conf *Configuration) PageByKey(c *gin.Context, key string) *Page {
	locale := context.GetLocale(c)

	if localizedPage, ok := conf.preprocessedPages[key][locale]; ok {
		page := localizedPage.Copy()
		page.InitializeData(c)
		return page
	}

	return nil
}

// Page : returns the standard page struct with the key and locale.
func (conf *Configuration) Page(c *gin.Context) *Page {
	key := context.GetPageKey(c)

	return conf.PageByKey(c, key)
}

// IsPartial : returns true if the page is a partial page.
func (conf *Configuration) IsPartial(key, locale string) bool {
	if localizedPage, ok := conf.preprocessedPages[key][locale]; ok {
		return localizedPage.Partial
	}

	return true
}

// URL : returns the URL of a page with its locale and some vars.
func (conf *Configuration) URL(key, locale string) string {
	url := conf.urls[key][locale]
	return url
}

// AsyncURL : returns the URL of a async url.
func (conf *Configuration) AsyncURL(key, locale string) string {
	if async, ok := conf.Async[key]; ok {
		return async.Paths[locale]
	}

	return ""
}
