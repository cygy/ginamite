package sitemap

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/worker/tasks"
	"github.com/sirupsen/logrus"
)

// Generate : generates the sitemaps.
func Generate(taskName string, getLocalizedVariables func(routeKey string, offset, limit int) map[string][]map[string]string) {
	file := config.Main.Sitemap.Name
	path := config.Main.Sitemap.Path
	staticHost := config.Main.Hosts.Static
	locales := config.Main.SupportedLocales

	// Create the index sitemap.
	filename := fmt.Sprintf("%s_index", file)
	indexSitemap, err := NewFile(path, filename, true)
	if err != nil {
		tasks.LogError(taskName, err, &logrus.Fields{
			"filename": filename,
			"path":     path,
		})
		return
	}

	// Define a closure to close and add multiple sitemaps to the index sitemap.
	closeSitemaps := func(sitemaps map[string]*File) {
		for _, sitemap := range sitemaps {
			if err := sitemap.Close(); err != nil {
				tasks.LogError(taskName, err, &logrus.Fields{
					"filename": sitemap.Name,
					"path":     sitemap.Path,
				})
				continue
			}

			sitemapURL := fmt.Sprintf("%s/%s", path, sitemap.Name)
			sitemapURL = filepath.Clean(sitemapURL)
			sitemapURL = fmt.Sprintf("%s%s", staticHost, sitemapURL)
			indexSitemap.AddSitemap(sitemapURL, time.Now())
		}
	}

	// Create the main sitemaps.
	mainSitemaps := map[string]*File{}
	for _, locale := range locales {
		filename := fmt.Sprintf("%s_%s_main", file, locale)
		sitemap, err := NewFile(path, filename, false)
		if err != nil {
			tasks.LogError(taskName, err, &logrus.Fields{
				"filename": filename,
				"path":     path,
			})
			continue
		}

		mainSitemaps[locale] = sitemap
	}

	// Add the urls.
	for key, route := range weburl.Main.Routes {
		if !route.Sitemap.Included {
			continue
		}

		if route.HasVariable() {
			// Define a closure to create multiple sitemaps per locale.
			round := 0
			createSitemaps := func(index int) map[string]*File {
				sitemaps := map[string]*File{}
				for _, locale := range locales {
					filename := fmt.Sprintf("%s_%s_%s_%d", file, locale, key, index)
					sitemap, err := NewFile(path, filename, false)
					if err != nil {
						tasks.LogError(taskName, err, &logrus.Fields{
							"filename": filename,
							"path":     path,
						})
						continue
					}

					sitemaps[locale] = sitemap
				}

				return sitemaps
			}

			// Create a sitemap per locale.
			sitemaps := createSitemaps(round)

			// The urls are added by batch of 500.
			offset := 0
			limit := 500
			for {
				localizedVariables := getLocalizedVariables(key, offset, limit)
				if len(localizedVariables) == 0 {
					break
				}

				// Is the limit of urls per file reached? If yes close the current sitemaps and create new ones.
				if (offset+limit)/config.Main.Sitemap.MaximumURLsPerPage > round {
					round++
					closeSitemaps(sitemaps)
					sitemaps = createSitemaps(round)
				}

				for locale, arrayOfVariables := range localizedVariables {
					absolutePath, isPathPresent := route.AbsolutePaths[locale]
					sitemap, isSitemapPresent := sitemaps[locale]
					if !isPathPresent || !isSitemapPresent || !config.Main.SupportsLocale(locale) {
						continue
					}

					for _, variables := range arrayOfVariables {
						path := absolutePath
						for key, value := range variables {
							path = strings.Replace(path, ":"+key, value, 1)
						}

						sitemap.AddURL(path, route.Sitemap.UpdateFrequency.String(), route.Sitemap.Priority)
					}
				}

				offset += limit
			}

			// Close the sitemaps.
			closeSitemaps(sitemaps)
		} else {
			// Add the urls to the main sitemap.
			for _, locale := range locales {
				path, ok := route.AbsolutePaths[locale]
				if !ok {
					continue
				}

				if sitemap, ok := mainSitemaps[locale]; ok {
					sitemap.AddURL(path, route.Sitemap.UpdateFrequency.String(), route.Sitemap.Priority)
				}
			}
		}
	}

	// Close the sitemaps.
	closeSitemaps(mainSitemaps)
	indexSitemap.Close()

	tasks.LogDone(taskName, nil)
}
