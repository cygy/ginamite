// Package image saves, resizes and deletes the images uploaded to the API.
package image

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// SaveFill saves different versions of an image to destination.
func SaveFill(src, directoryDest string, width, height int, saveSmallVersion, saveThumbVersion bool) (*Versions, error) {
	return save(src, directoryDest, width, height, saveSmallVersion, saveThumbVersion, false)
}

// SaveFit saves different versions of an image to destination.
func SaveFit(src, directoryDest string, width, height int, saveSmallVersion, saveThumbVersion bool) (*Versions, error) {
	return save(src, directoryDest, width, height, saveSmallVersion, saveThumbVersion, true)
}

// Delete deletes some images.
func Delete(filenames []string, directory string) {
	for _, filename := range filenames {
		if len(filename) > 0 && len(directory) > 0 {
			path := directory + "/" + filename
			err := os.Remove(path)
			if err != nil {
				log.WithFields(logrus.Fields{
					"path":  path,
					"error": err.Error(),
				}).Error("unable to delete a file")
			}
		}
	}
}

// NewVersions returns a new NewVersions struct.
func NewVersions(src string, width, height int, directoryDest string, saveSmallVersion, saveThumbVersion bool) *Versions {
	ext := strings.ToLower(filepath.Ext(src))
	filename := strings.Replace(strings.ToLower(filepath.Base(src)), ext, "", 1)

	versions := &Versions{}
	versions.Full.Size1x.Width = width
	versions.Full.Size1x.Height = height
	versions.Full.Size1x.DefineNameAndPaths(filename, ext, suffixFull, suffix1x, directoryDest)
	versions.Full.Size2x.Width = width * 2
	versions.Full.Size2x.Height = height * 2
	versions.Full.Size2x.DefineNameAndPaths(filename, ext, suffixFull, suffix2x, directoryDest)

	if saveSmallVersion {
		baseWidth := width / 3
		baseHeight := height / 3

		versions.Small.Size1x.Width = baseWidth
		versions.Small.Size1x.Height = baseHeight
		versions.Small.Size1x.DefineNameAndPaths(filename, ext, suffixSmall, suffix1x, directoryDest)
		versions.Small.Size2x.Width = baseWidth * 2
		versions.Small.Size2x.Height = baseHeight * 2
		versions.Small.Size2x.DefineNameAndPaths(filename, ext, suffixSmall, suffix2x, directoryDest)
	}

	if saveThumbVersion {
		baseWidth := width / 6
		baseHeight := height / 6

		versions.Thumb.Size1x.Width = baseWidth
		versions.Thumb.Size1x.Height = baseHeight
		versions.Thumb.Size1x.DefineNameAndPaths(filename, ext, suffixThumb, suffix1x, directoryDest)
		versions.Thumb.Size2x.Width = baseWidth * 2
		versions.Thumb.Size2x.Height = baseHeight * 2
		versions.Thumb.Size2x.DefineNameAndPaths(filename, ext, suffixThumb, suffix2x, directoryDest)
	}

	return versions
}
