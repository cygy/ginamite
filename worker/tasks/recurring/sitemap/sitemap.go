// Package sitemap generates the XML sitemaps of the web routes.
package sitemap

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/worker/tasks"

	"github.com/sirupsen/logrus"
)

// batchSize is the number of variables requested at once to generate the URLs of a route.
const batchSize = 500

// GetLocalizedVariablesFunc returns the variables of the URLs of a route, from offset, up to limit, by locale.
// An empty result means there are no more variables.
type GetLocalizedVariablesFunc = func(routeKey string, offset, limit int) map[string][]map[string]string

// generator generates the sitemaps of the routes and their index.
type generator struct {
	taskName       string
	name           string // The prefix of the files.
	path           string // The directory of the files.
	staticHost     string
	locales        []string
	supportsLocale func(locale string) bool
	routes         *weburl.Configuration
	maxURLs        int // The maximum number of URLs per file of a route with variables.
	getVariables   GetLocalizedVariablesFunc
	index          *File
}

// Generate generates the sitemaps of the routes included in the sitemap, and their index, in the directory
// of the configuration, which is given. The routes without variables are in a sitemap per locale. The routes with variables, whose
// values come from getLocalizedVariables, have one sitemap per locale and per maximum number of URLs.
// The errors are logged.
func Generate(taskName string, configuration *config.Configuration, routes *weburl.Configuration, getLocalizedVariables GetLocalizedVariablesFunc) {
	g := generator{
		taskName:       taskName,
		name:           configuration.Sitemap.Name,
		path:           configuration.Sitemap.Path,
		staticHost:     configuration.Hosts.Static,
		locales:        configuration.SupportedLocales,
		supportsLocale: configuration.SupportsLocale,
		routes:         routes,
		maxURLs:        configuration.Sitemap.MaximumURLsPerPage,
		getVariables:   getLocalizedVariables,
	}

	g.generate()
}

// generate generates the sitemaps and their index.
func (g *generator) generate() {
	index, ok := g.newFile(fmt.Sprintf("%s_index", g.name), true)
	if !ok {
		return
	}
	g.index = index

	// The URLs of the routes without variables are in the main sitemaps.
	mainSitemaps := map[string]*File{}
	for _, locale := range g.locales {
		if sitemap, ok := g.newFile(fmt.Sprintf("%s_%s_main", g.name, locale), false); ok {
			mainSitemaps[locale] = sitemap
		}
	}

	for key, route := range g.routes.Routes {
		if !route.Sitemap.Included {
			continue
		}

		if route.HasVariable() {
			g.addRouteWithVariables(key, route)
			continue
		}

		for locale, sitemap := range mainSitemaps {
			if path, ok := route.AbsolutePaths[locale]; ok {
				sitemap.AddURL(path, route.Sitemap.UpdateFrequency.String(), route.Sitemap.Priority)
			}
		}
	}

	g.closeSitemaps(mainSitemaps)
	if err := g.index.Close(); err != nil {
		log.WithField("error", err).Error("unable to close the index sitemap")
	}

	tasks.LogDone(g.taskName, nil)
}

// addRouteWithVariables adds the URLs of a route with variables, by batch, to sitemaps which are
// closed and replaced by new ones when the maximum number of URLs is reached.
func (g *generator) addRouteWithVariables(key string, route weburl.Route) {
	round := 0
	sitemaps := g.createSitemaps(key, round)

	for offset := 0; ; offset += batchSize {
		localizedVariables := g.getVariables(key, offset, batchSize)
		if len(localizedVariables) == 0 {
			break
		}

		// Is the limit of URLs per file reached? If yes, close the current sitemaps and create new ones.
		if g.maxURLs > 0 && (offset+batchSize)/g.maxURLs > round {
			round++
			g.closeSitemaps(sitemaps)
			sitemaps = g.createSitemaps(key, round)
		}

		for locale, arrayOfVariables := range localizedVariables {
			absolutePath, isPathPresent := route.AbsolutePaths[locale]
			sitemap, isSitemapPresent := sitemaps[locale]
			if !isPathPresent || !isSitemapPresent || !g.supportsLocale(locale) {
				continue
			}

			for _, variables := range arrayOfVariables {
				sitemap.AddURL(expandPath(absolutePath, variables), route.Sitemap.UpdateFrequency.String(), route.Sitemap.Priority)
			}
		}
	}

	g.closeSitemaps(sitemaps)
}

// createSitemaps creates a sitemap per locale for the round of a route.
func (g *generator) createSitemaps(key string, round int) map[string]*File {
	sitemaps := map[string]*File{}
	for _, locale := range g.locales {
		if sitemap, ok := g.newFile(fmt.Sprintf("%s_%s_%s_%d", g.name, locale, key, round), false); ok {
			sitemaps[locale] = sitemap
		}
	}

	return sitemaps
}

// closeSitemaps closes the sitemaps and adds the ones which are closed without error to the index.
func (g *generator) closeSitemaps(sitemaps map[string]*File) {
	for _, sitemap := range sitemaps {
		if err := sitemap.Close(); err != nil {
			tasks.LogError(g.taskName, err, &logrus.Fields{
				"filename": sitemap.Name,
				"path":     sitemap.Path,
			})
			continue
		}

		sitemapURL := g.staticHost + filepath.Clean(fmt.Sprintf("%s/%s", g.path, sitemap.Name))
		g.index.AddSitemap(sitemapURL, time.Now())
	}
}

// newFile creates a sitemap file, or logs the error and returns false.
func (g *generator) newFile(filename string, isIndex bool) (*File, bool) {
	sitemap, err := NewFile(g.path, filename, isIndex)
	if err != nil {
		tasks.LogError(g.taskName, err, &logrus.Fields{
			"filename": filename,
			"path":     g.path,
		})
		return nil, false
	}

	return sitemap, true
}

// expandPath replaces the variables ":name" of a path by their value. The longest names are replaced first, so that
// ":id" does not replace the start of ":identifier".
func expandPath(path string, variables map[string]string) string {
	names := make([]string, 0, len(variables))
	for name := range variables {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		if len(names[i]) != len(names[j]) {
			return len(names[i]) > len(names[j])
		}

		return names[i] < names[j]
	})

	for _, name := range names {
		path = strings.Replace(path, ":"+name, variables[name], 1)
	}

	return path
}
