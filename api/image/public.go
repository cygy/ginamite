package image

import (
	"os"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// SaveFill : saves different versions of an image to destination.
func SaveFill(src, directoryDest string, width, height int, saveSmallVersion, saveThumbVersion bool) (*Versions, error) {
	return save(src, directoryDest, width, height, saveSmallVersion, saveThumbVersion, false)
}

// SaveFit : saves different versions of an image to destination.
func SaveFit(src, directoryDest string, width, height int, saveSmallVersion, saveThumbVersion bool) (*Versions, error) {
	return save(src, directoryDest, width, height, saveSmallVersion, saveThumbVersion, true)
}

// Delete : deletes some images.
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
