package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSave_UnwritableDestination(t *testing.T) {
	src := writePNG(t, t.TempDir(), "photo.png", 20, 20)

	// The destination is a regular file: neither the directories nor the images can be created.
	dest := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(dest, nil, 0o600); err != nil {
		t.Fatalf("unable to create the destination file: %v", err)
	}

	for _, fit := range []bool{true, false} {
		if _, err := save(src, dest, 10, 10, false, false, fit); err == nil {
			t.Errorf("save(fit=%v) error = nil, want an error for an unwritable destination", fit)
		}
	}
}
