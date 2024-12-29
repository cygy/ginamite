package weburl

import (
	"fmt"
	"os"
	"strings"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/web/html"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

// GetURL : returns the full URL of a web route.
func (conf *Configuration) GetURL(key, locale string) string {
	url, ok := conf.Routes[key].AbsolutePaths[locale]
	if !ok {
		log.WithFields(logrus.Fields{
			"route":  key,
			"locale": locale,
		}).Warn("web url not found")
	}

	return url
}

// Initialize : initializes the configuration of the routes.
func (conf *Configuration) Initialize(file, webHost string) bool {
	// Read the configuration file.
	source, err := os.ReadFile(file)
	if err != nil {
		log.WithFields(logrus.Fields{
			"path":  file,
			"error": err.Error(),
		}).Panic("unable to load web urls")
		return false
	}

	// Get the routes from the file.
	err = yaml.Unmarshal(source, conf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"path":  file,
			"error": err.Error(),
		}).Panic("unable to load web urls")
		return false
	}

	// Normalize the routes.
	localeKey := fmt.Sprintf("{{.%s}}", html.LocaleKeyword)
	for k, r := range conf.Routes {
		absolutePaths := map[string]string{}

		for locale, path := range r.RelativePaths {
			url := fmt.Sprintf("%s%s", webHost, path)
			url = strings.Replace(url, localeKey, locale, -1)
			absolutePaths[locale] = url

			parts := strings.Split(locale, "-")
			if len(parts) == 2 {
				absolutePaths[parts[0]] = url
			}
		}

		r.AbsolutePaths = absolutePaths
		conf.Routes[k] = r
	}

	return true
}

// HasVariable : returns true if the route is a generic path, is it contains some variables.
func (route *Route) HasVariable() bool {
	hasVariable := false
	for _, path := range route.RelativePaths {
		hasVariable = strings.Contains(path, "/:")
		break
	}

	return hasVariable
}
