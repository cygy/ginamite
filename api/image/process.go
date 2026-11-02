package image

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/cygy/ginamite/common/log"

	"github.com/disintegration/imaging"
)

func save(src, directoryDest string, width, height int, saveSmallVersion, saveThumbVersion, fit bool) (*Versions, error) {
	if err := checkPixels(src); err != nil {
		log.Error("unable to accept a file image",
			"path", src,
			"error", err.Error(),
		)
		return nil, err
	}

	f, err := imaging.Open(src)
	if err != nil {
		log.Error("unable to open a file image",
			"path", src,
			"error", err.Error(),
		)
		return nil, err
	}

	versions := NewVersions(src, width, height, directoryDest, saveSmallVersion, saveThumbVersion)
	properties := versions.All()

	for _, property := range properties {
		// create the destination directory.
		directory := filepath.Dir(property.AbsolutePath)
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			if err := os.MkdirAll(directory, 0o750); err != nil {
				log.Error("unable to create directory",
					"path", directory,
					"error", err.Error(),
				)
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
			log.Error("unable to save a file image",
				"src", src,
				"dest", property.AbsolutePath,
				"error", err.Error(),
			)
			return nil, err
		}
	}

	return versions, nil
}

// checkPixels returns ErrTooManyPixels if the image src declares more pixels than MaxPixels, or an error if its header
// can not be read.
func checkPixels(src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error("unable to close a file image", "path", src, "error", err.Error())
		}
	}()

	config, _, err := image.DecodeConfig(f)
	if err != nil {
		return fmt.Errorf("unable to read the header of the image: %w", err)
	}

	if int64(config.Width)*int64(config.Height) > MaxPixels {
		return fmt.Errorf("%w: %dx%d", ErrTooManyPixels, config.Width, config.Height)
	}

	return nil
}
