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

func TestProperty(t *testing.T) {
	var p Property
	if p.IsDefined() {
		t.Error("an empty property must not be defined")
	}

	p.DefineNameAndPaths("photo", ".png", "-small", "@2x", "/dest")

	if want := "photo-small@2x.png"; p.Name != want {
		t.Errorf("Name = %q, want %q", p.Name, want)
	}
	if want := "p/h/photo-small@2x.png"; p.RelativePath != want {
		t.Errorf("RelativePath = %q, want %q", p.RelativePath, want)
	}
	if want := "/dest/p/h/photo-small@2x.png"; p.AbsolutePath != want {
		t.Errorf("AbsolutePath = %q, want %q", p.AbsolutePath, want)
	}
	if !p.IsDefined() {
		t.Error("a property with a name must be defined")
	}
}

func TestSizes_IsDefined(t *testing.T) {
	tests := []struct {
		name  string
		sizes Sizes
		want  bool
	}{
		{"none", Sizes{}, false},
		{"only 1x", Sizes{Size1x: Property{Name: "a"}}, false},
		{"only 2x", Sizes{Size2x: Property{Name: "a"}}, false},
		{"both", Sizes{Size1x: Property{Name: "a"}, Size2x: Property{Name: "b"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sizes.IsDefined(); got != tt.want {
				t.Errorf("IsDefined() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewVersions(t *testing.T) {
	tests := []struct {
		name            string
		small, thumb    bool
		wantCount       int
		wantSmallWidth  int
		wantThumbWidth  int
		wantFullName    string
		wantFull2xWidth int
		wantFull2xPath  string
	}{
		{"full only", false, false, 2, 0, 0, "photo.png", 600, "/dest/p/h/photo@2x.png"},
		{"with small", true, false, 4, 100, 0, "photo.png", 600, "/dest/p/h/photo@2x.png"},
		{"with thumb", false, true, 4, 0, 50, "photo.png", 600, "/dest/p/h/photo@2x.png"},
		{"with all", true, true, 6, 100, 50, "photo.png", 600, "/dest/p/h/photo@2x.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := NewVersions("/src/Photo.PNG", 300, 150, "/dest", tt.small, tt.thumb)

			if got := len(versions.All()); got != tt.wantCount {
				t.Errorf("len(All()) = %d, want %d", got, tt.wantCount)
			}
			if versions.Full.Size1x.Name != tt.wantFullName {
				t.Errorf("Full.Size1x.Name = %q, want %q", versions.Full.Size1x.Name, tt.wantFullName)
			}
			if versions.Full.Size2x.Width != tt.wantFull2xWidth || versions.Full.Size2x.AbsolutePath != tt.wantFull2xPath {
				t.Errorf("Full.Size2x = %+v, want width %d and path %q", versions.Full.Size2x, tt.wantFull2xWidth, tt.wantFull2xPath)
			}
			if versions.Small.Size1x.Width != tt.wantSmallWidth || versions.Thumb.Size1x.Width != tt.wantThumbWidth {
				t.Errorf("small, thumb width = %d, %d, want %d, %d", versions.Small.Size1x.Width, versions.Thumb.Size1x.Width, tt.wantSmallWidth, tt.wantThumbWidth)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name string
		save func(src, dest string, w, h int, small, thumb bool) (*Versions, error)
	}{
		{"fill", SaveFill},
		{"fit", SaveFit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := writePNG(t, t.TempDir(), "photo.png", 120, 60)
			dest := t.TempDir()

			versions, err := tt.save(src, dest, 60, 30, true, true)
			if err != nil {
				t.Fatalf("save error: %v", err)
			}

			properties := versions.All()
			if len(properties) != 6 {
				t.Fatalf("len(All()) = %d, want 6", len(properties))
			}
			for _, p := range properties {
				if w, h := imageSize(t, p.AbsolutePath); w != p.Width || h != p.Height {
					t.Errorf("%s size = %dx%d, want %dx%d", p.Name, w, h, p.Width, p.Height)
				}
			}
		})
	}
}

func TestSave_UnreadableSource(t *testing.T) {
	if _, err := SaveFill(filepath.Join(t.TempDir(), "missing.png"), t.TempDir(), 10, 10, false, false); err == nil {
		t.Error("SaveFill() error = nil, want an error for a missing source")
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	first := writePNG(t, dir, "first.png", 2, 2)
	second := writePNG(t, dir, "second.png", 2, 2)

	// Missing files and empty arguments are ignored.
	Delete([]string{"first.png", "", "missing.png"}, dir)
	Delete([]string{"second.png"}, "")

	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Error("first.png must be deleted")
	}
	if _, err := os.Stat(second); err != nil {
		t.Error("second.png must be kept when the directory is empty")
	}
}
