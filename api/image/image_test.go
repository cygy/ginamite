package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveFillAndFit(t *testing.T) {
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

func TestSaveFill_UnreadableSource(t *testing.T) {
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
