package image

import "errors"

// MaxPixels is the maximum number of pixels of an image saved, 50 megapixels, read from its header before it is
// decoded: a small file can declare huge dimensions and exhaust the memory once decoded.
const MaxPixels = 50_000_000

// ErrTooManyPixels is returned when an image has more pixels than MaxPixels.
var ErrTooManyPixels = errors.New("the image has too many pixels")

const (
	suffix1x    = ""
	suffix2x    = "@2x"
	suffixFull  = ""
	suffixSmall = "-small"
	suffixThumb = "-thumb"
)
