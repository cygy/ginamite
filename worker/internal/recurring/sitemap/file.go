package sitemap

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/random"
)

// File is a sitemap file, gzipped. It is written to a temporary file, moved to its path when closed.
type File struct {
	Name     string
	Path     string
	tempPath string
	writer   *gzip.Writer
	file     *os.File
	isIndex  bool
	err      error // The first error of the writes.
}

// NewFile creates the sitemap file "<name>.xml.gz" in the directory, and writes its header.
// The file is an index of sitemaps, and not a sitemap, if isIndex is true.
func NewFile(directoryPath, name string, isIndex bool) (*File, error) {
	filename := name + ".xml.gz"

	f := File{
		Name:     filename,
		Path:     filepath.Clean(directoryPath + "/" + filename),
		tempPath: filepath.Clean(directoryPath + "/" + random.String(48)),
		isIndex:  isIndex,
	}

	fo, err := os.Create(f.tempPath)
	if err != nil {
		f.logError("create sitemap file", err)
		return nil, err
	}

	f.file = fo
	f.writer = gzip.NewWriter(fo)

	if isIndex {
		f.write("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<sitemapindex xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	} else {
		f.write("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	}

	return &f, nil
}

// AddURL adds an URL to the sitemap. It does nothing for an index.
func (f *File) AddURL(url, updateFrequency string, priority float32) {
	if f.isIndex {
		return
	}

	f.write(fmt.Sprintf("<url>\n<loc>%s</loc>\n<changefreq>%s</changefreq>\n<priority>%.1f</priority>\n</url>\n", escape(url), escape(updateFrequency), priority))
}

// AddSitemap adds a reference to a sitemap file to the index. It does nothing for a sitemap.
func (f *File) AddSitemap(url string, lastModified time.Time) {
	if !f.isIndex {
		return
	}

	// The date format must be: 2006-01-02T15:04:05-07:00
	f.write(fmt.Sprintf("<sitemap>\n<loc>%s</loc>\n<lastmod>%s</lastmod>\n</sitemap>\n", escape(url), lastModified.Format("2006-01-02T15:04:05-07:00")))
}

// Close writes the end of the file, closes it and moves it to its path.
// The file is deleted, and the error returned, if it could not be written entirely.
func (f *File) Close() error {
	if f.isIndex {
		f.write("</sitemapindex>")
	} else {
		f.write("</urlset>")
	}

	err := f.err
	if closeErr := f.writer.Close(); closeErr != nil {
		f.logError("close sitemap writer", closeErr)
		err = firstError(err, closeErr)
	}
	if closeErr := f.file.Close(); closeErr != nil {
		f.logError("close sitemap file", closeErr)
		err = firstError(err, closeErr)
	}

	if err == nil {
		if err = os.Rename(f.tempPath, f.Path); err != nil {
			f.logError("move temp sitemap file", err)
		}
	}

	if err != nil {
		_ = os.Remove(f.tempPath)
	}

	return err
}

// write writes s to the file. The first error is logged and returned by Close.
func (f *File) write(s string) {
	if _, err := f.writer.Write([]byte(s)); err != nil {
		f.logError("write to sitemap file", err)
		f.err = firstError(f.err, err)
	}

	if err := f.writer.Flush(); err != nil {
		f.logError("flush to sitemap file", err)
		f.err = firstError(f.err, err)
	}
}

// logError logs an error about the file.
func (f *File) logError(message string, err error) {
	log.Error(message,
		"path", f.Path,
		"tempPath", f.tempPath,
		"error", err,
	)
}

// firstError returns the first non-nil error.
func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// escape returns s escaped to be a text of an XML element.
func escape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s)) // Writing to a buffer never fails.

	return b.String()
}
