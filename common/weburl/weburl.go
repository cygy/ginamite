// Package weburl resolves the URLs of the web routes from their key and their locale.
package weburl

import (
	"fmt"
	"os"
	"strings"

	"github.com/cygy/ginamite/common/log"

	yaml "gopkg.in/yaml.v3"
)

// Configuration holds the web routes, loaded from a YAML file.
type Configuration struct {
	Routes map[string]Route `yaml:"routes"` // The index is the unique key of the route (i.e. "account.registration").
}

// Route is a web route, with its path for each locale.
type Route struct {
	RelativePaths map[string]string `yaml:"paths"`
	AbsolutePaths map[string]string `yaml:"-"` // The index is the locale of the route (i.e. "fr-fr"), the value is the localized route (i.e. "http://.../account/registration").
	Sitemap       Sitemap           `yaml:"sitemap"`
}

// NewConfiguration returns an empty Configuration.
func NewConfiguration() (conf *Configuration) {
	conf = &Configuration{
		Routes: map[string]Route{},
	}

	return
}

// GetURL returns the absolute URL of the web route key for the locale.
// It logs a warning and returns an empty string if the route or the locale is unknown.
func (conf *Configuration) GetURL(key, locale string) string {
	url, ok := conf.Routes[key].AbsolutePaths[locale]
	if !ok {
		log.Warn("web url not found",
			"route", key,
			"locale", locale,
		)
	}

	return url
}

// Initialize loads the routes from the YAML file and computes their absolute paths,
// prefixing the paths with webHost and replacing the locale keyword.
// A route is also reachable with the language of a locale ("fr" for "fr-fr").
// It returns an error if the file can not be read or parsed.
func (conf *Configuration) Initialize(file, webHost string) error {
	source, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("unable to read the routes: %w", err)
	}

	if err := yaml.Unmarshal(source, conf); err != nil {
		return fmt.Errorf("unable to load the routes %s: %w", file, err)
	}

	// Normalize the routes.
	localeKey := fmt.Sprintf("{{.%s}}", LocaleKeyword)
	for k, r := range conf.Routes {
		absolutePaths := map[string]string{}

		for locale, path := range r.RelativePaths {
			url := fmt.Sprintf("%s%s", webHost, path)
			url = strings.ReplaceAll(url, localeKey, locale)
			absolutePaths[locale] = url

			parts := strings.Split(locale, "-")
			if len(parts) == 2 {
				absolutePaths[parts[0]] = url
			}
		}

		r.AbsolutePaths = absolutePaths
		conf.Routes[k] = r
	}

	return nil
}

// HasVariable reports whether a path of the route is generic, i.e. contains some variables.
func (route *Route) HasVariable() bool {
	for _, path := range route.RelativePaths {
		if strings.Contains(path, "/:") {
			return true
		}
	}

	return false
}

// Load returns the configuration of the web routes read from the YAML file, with the absolute paths computed.
// See Configuration.Initialize.
func Load(file, webHost string) (*Configuration, error) {
	conf := NewConfiguration()
	if err := conf.Initialize(file, webHost); err != nil {
		return nil, err
	}

	return conf, nil
}
