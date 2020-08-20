package sitemap

import (
	"fmt"
	"os"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/sirupsen/logrus"
)

// AddURL : adds an URL to the sitemap.
func (f *File) AddURL(url, updateFrequency string, priority float32) {
	if f.isIndex {
		return
	}

	s := fmt.Sprintf("<url>\n<loc>%s</loc>\n<changefreq>%s</changefreq>\n<priority>%.1f</priority>\n</url>\n", url, updateFrequency, priority)
	f.write(s)
}

// AddSitemap : adds a reference to a sitemap file to the index.
func (f *File) AddSitemap(url string, lastModified time.Time) {
	if !f.isIndex {
		return
	}

	// Date format must be : 2006-01-02T15:04:05-07:00
	s := fmt.Sprintf("<sitemap>\n<loc>%s</loc>\n<lastmod>%s</lastmod>\n</sitemap>", url, lastModified.Format("2006-01-02T15:04:05-07:00"))
	f.write(s)
}

// Close : closes the file.
func (f *File) Close() error {
	if f.isIndex {
		f.write("</sitemapindex>")
	} else {
		f.write("</urlset>")
	}

	if err := f.writer.Close(); err != nil {
		log.WithFields(logrus.Fields{
			"path":     f.Path,
			"tempPath": f.tempPath,
			"error":    err,
		}).Error("close sitemap writer")
	}

	if err := f.file.Close(); err != nil {
		log.WithFields(logrus.Fields{
			"path":     f.Path,
			"tempPath": f.tempPath,
			"error":    err,
		}).Error("close sitemap file")
	}

	if err := os.Rename(f.tempPath, f.Path); err != nil {
		log.WithFields(logrus.Fields{
			"path":     f.Path,
			"tempPath": f.tempPath,
			"error":    err,
		}).Error("move temp sitemap file")
		return err
	}

	return nil
}

// Private functions.
func (f *File) write(s string) {
	if _, err := f.writer.Write([]byte(s)); err != nil {
		log.WithFields(logrus.Fields{
			"path":     f.Path,
			"tempPath": f.tempPath,
			"error":    err,
		}).Error("write to sitemap file")
	}

	if err := f.writer.Flush(); err != nil {
		log.WithFields(logrus.Fields{
			"path":     f.Path,
			"tempPath": f.tempPath,
			"error":    err,
		}).Error("flush to sitemap file")
	}
}
