package weburl

import "github.com/cygy/ginamite/web/html"

// Configuration : struct of the configuration of the routes.
type Configuration struct {
	Routes map[string]Route `yaml:"routes"` // The index is the unique key of the route (i.e. "account.registration").
}

// Route : struct of a web route
type Route struct {
	RelativePaths map[string]string `yaml:"paths"`
	AbsolutePaths map[string]string `yaml:"-"` // The index is the locale of the route (i.e. "fr-fr"), the value is the localized route (i.e. "http://.../account/registration").
	Sitemap       html.Sitemap      `yaml:"sitemap"`
}

// NewConfiguration : returns a Configuration struct.
func NewConfiguration() (conf *Configuration) {
	conf = &Configuration{
		Routes: map[string]Route{},
	}

	return
}
