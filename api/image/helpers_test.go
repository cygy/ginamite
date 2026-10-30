package image

import (
	goimage "image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// writePNG writes a width x height PNG file and returns its path.
func writePNG(t *testing.T, dir, name string, width, height int) string {
	t.Helper()

	img := goimage.NewNRGBA(goimage.Rect(0, 0, width, height))
	for x := range width {
		for y := range height {
			img.Set(x, y, color.NRGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("unable to create the image: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("unable to close the image: %v", err)
		}
	}()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("unable to encode the image: %v", err)
	}

	return path
}

func imageSize(t *testing.T, path string) (int, int) {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("unable to open the image: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("unable to close the image: %v", err)
		}
	}()

	cfg, _, err := goimage.DecodeConfig(f)
	if err != nil {
		t.Fatalf("unable to decode the image: %v", err)
	}

	return cfg.Width, cfg.Height
}
