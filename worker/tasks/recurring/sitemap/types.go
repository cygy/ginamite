package sitemap

import (
	"compress/gzip"
	"os"
	"path/filepath"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/random"
	"github.com/sirupsen/logrus"
)

// File : struct defining the properties of a sitemap file.
type File struct {
	Name     string
	Path     string
	tempPath string
	writer   *gzip.Writer
	file     *os.File
	isIndex  bool
}

// NewFile : creates and returns a new struct 'File'.
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
		log.WithFields(logrus.Fields{
			"path":     f.Path,
			"tempPath": f.tempPath,
			"error":    err,
		}).Error("create sitemap file")
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
