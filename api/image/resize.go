package image

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"
)

// jpegQuality is the quality of the JPEG images saved.
const jpegQuality = 95

// ErrUnsupportedFormat is returned when an image is saved with an extension whose format is unknown.
var ErrUnsupportedFormat = errors.New("unsupported image format")

// open decodes the image of the file src, in one of the formats registered: JPEG, PNG, GIF, BMP or TIFF.
func open(src string) (img image.Image, err error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()

	img, _, err = image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("unable to decode the image: %w", err)
	}

	return img, nil
}

// fit returns the image scaled down, keeping its proportions, to fit in width x height. An image which fits already
// is returned as is.
func fit(img image.Image, width, height int) image.Image {
	b := img.Bounds()
	if b.Dx() <= width && b.Dy() <= height {
		return img
	}

	// Scale by the smaller ratio, comparing width/b.Dx() and height/b.Dy() without floats.
	w, h := width, b.Dy()*width/b.Dx()
	if b.Dx()*height < b.Dy()*width {
		w, h = b.Dx()*height/b.Dy(), height
	}

	return scale(img, image.Rect(0, 0, max(w, 1), max(h, 1)), b)
}

// fill returns the image scaled, keeping its proportions, to cover width x height, then cropped at its center to
// width x height.
func fill(img image.Image, width, height int) image.Image {
	b := img.Bounds()

	// The part of the source with the proportions of the destination, at the center.
	crop := b
	if b.Dx()*height > b.Dy()*width {
		w := b.Dy() * width / height
		crop.Min.X += (b.Dx() - w) / 2
		crop.Max.X = crop.Min.X + w
	} else {
		h := b.Dx() * height / width
		crop.Min.Y += (b.Dy() - h) / 2
		crop.Max.Y = crop.Min.Y + h
	}

	return scale(img, image.Rect(0, 0, width, height), crop)
}

// scale draws the part src of img in an image of the size of dst, with the Catmull-Rom filter.
func scale(img image.Image, dst, src image.Rectangle) image.Image {
	out := image.NewNRGBA(dst)
	draw.CatmullRom.Scale(out, dst, img, src, draw.Src, nil)

	return out
}

// write saves the image to the file path, in the format of its extension: .jpg or .jpeg, .png, .gif, .bmp, .tif or
// .tiff. It returns ErrUnsupportedFormat for another extension, without creating the file.
func write(img image.Image, path string) (err error) {
	encode, err := encoder(path)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()

	return encode(f, img)
}

// encoder returns the function encoding an image in the format of the extension of path.
func encoder(path string) (func(io.Writer, image.Image) error, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return func(w io.Writer, img image.Image) error {
			return jpeg.Encode(w, img, &jpeg.Options{Quality: jpegQuality})
		}, nil
	case ".png":
		return png.Encode, nil
	case ".gif":
		return func(w io.Writer, img image.Image) error { return gif.Encode(w, img, nil) }, nil
	case ".bmp":
		return bmp.Encode, nil
	case ".tif", ".tiff":
		return func(w io.Writer, img image.Image) error { return tiff.Encode(w, img, nil) }, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFormat, filepath.Ext(path))
	}
}
