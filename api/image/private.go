package image

import (
	"image"
	"os"
	"path/filepath"

	"github.com/cygy/ginamite/common/log"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
)

func save(src, directoryDest string, width, height int, saveSmallVersion, saveThumbVersion, fit bool) (*Versions, error) {
	f, err := imaging.Open(src)
	if err != nil {
		log.WithFields(logrus.Fields{
			"path":  src,
			"error": err.Error(),
		}).Error("unable to open a file image")
		return nil, err
	}

	versions := NewVersions(src, width, height, directoryDest, saveSmallVersion, saveThumbVersion)
	properties := versions.All()

	for _, property := range properties {
		// create the destination directory.
		directory := filepath.Dir(property.AbsolutePath)
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			if err := os.MkdirAll(directory, os.ModePerm); err != nil {
				log.WithFields(logrus.Fields{
					"path":  directory,
					"error": err.Error(),
				}).Error("unable to create directory")
			}
		}

		var image *image.NRGBA

		if fit {
			image = imaging.Fit(f, property.Width, property.Height, imaging.Lanczos)
		} else {
			image = imaging.Fill(f, property.Width, property.Height, imaging.Center, imaging.Lanczos)
		}

		err = imaging.Save(image, property.AbsolutePath)
		if err != nil {
			log.WithFields(logrus.Fields{
				"src":   src,
				"dest":  property.AbsolutePath,
				"error": err.Error(),
			}).Error("unable to save a file image")
			return nil, err
		}
	}

	return versions, nil
}
