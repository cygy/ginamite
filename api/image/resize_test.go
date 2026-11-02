package image

import (
	"errors"
	goimage "image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestFit(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		boxW, boxH    int
		wantW, wantH  int
	}{
		{"landscape", 200, 100, 50, 50, 50, 25},
		{"portrait", 100, 200, 50, 50, 25, 50},
		{"already fits: not enlarged", 20, 10, 50, 50, 20, 10},
		{"thin line keeps a pixel", 1000, 1, 10, 10, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fit(goimage.NewNRGBA(goimage.Rect(0, 0, tt.width, tt.height)), tt.boxW, tt.boxH).Bounds()
			if got.Dx() != tt.wantW || got.Dy() != tt.wantH {
				t.Errorf("fit() = %dx%d, want %dx%d", got.Dx(), got.Dy(), tt.wantW, tt.wantH)
			}
		})
	}
}

func TestFill(t *testing.T) {
	// A wide image: left third red, middle third green, right third blue. Filling a square keeps the center.
	img := goimage.NewNRGBA(goimage.Rect(0, 0, 300, 100))
	for x := range 300 {
		c := []color.NRGBA{{R: 255, A: 255}, {G: 255, A: 255}, {B: 255, A: 255}}[x/100]
		for y := range 100 {
			img.SetNRGBA(x, y, c)
		}
	}

	got := fill(img, 10, 10)
	if b := got.Bounds(); b.Dx() != 10 || b.Dy() != 10 {
		t.Fatalf("fill() = %dx%d, want 10x10", b.Dx(), b.Dy())
	}
	if r, g, b, _ := got.At(5, 5).RGBA(); r != 0 || g == 0 || b != 0 {
		t.Errorf("center = %v, want the green center of the source", got.At(5, 5))
	}

	// A tall image is cropped vertically, and a small one enlarged.
	if b := fill(goimage.NewNRGBA(goimage.Rect(0, 0, 5, 20)), 10, 10).Bounds(); b.Dx() != 10 || b.Dy() != 10 {
		t.Errorf("fill(tall) = %dx%d, want 10x10", b.Dx(), b.Dy())
	}
}

func TestWriteAndOpen(t *testing.T) {
	img := goimage.NewNRGBA(goimage.Rect(0, 0, 4, 3))
	dir := t.TempDir()

	for _, ext := range []string{".jpg", ".JPEG", ".png", ".gif", ".bmp", ".tif", ".tiff"} {
		path := filepath.Join(dir, "image"+ext)
		if err := write(img, path); err != nil {
			t.Fatalf("write(%s) = %v", ext, err)
		}

		decoded, err := open(path)
		if err != nil {
			t.Fatalf("open(%s) = %v", ext, err)
		}
		if b := decoded.Bounds(); b.Dx() != 4 || b.Dy() != 3 {
			t.Errorf("%s: size = %dx%d, want 4x3", ext, b.Dx(), b.Dy())
		}
	}
}

func TestWrite_Errors(t *testing.T) {
	img := goimage.NewNRGBA(goimage.Rect(0, 0, 1, 1))
	dir := t.TempDir()

	path := filepath.Join(dir, "image.webp")
	if err := write(img, path); !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("write(.webp) = %v, want ErrUnsupportedFormat", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("no file must be created for an unsupported format")
	}

	if err := write(img, filepath.Join(dir, "missing", "image.png")); err == nil {
		t.Error("write() into a missing directory must fail")
	}
}

func TestOpen_Errors(t *testing.T) {
	if _, err := open(filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Error("open() of a missing file must fail")
	}

	path := filepath.Join(t.TempDir(), "text.png")
	if err := os.WriteFile(path, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := open(path); err == nil {
		t.Error("open() of a text file must fail")
	}
}
