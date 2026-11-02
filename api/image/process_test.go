package image

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
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

// withDimensions rewrites the dimensions declared in the header of the PNG file path, without changing its pixels.
func withDimensions(t *testing.T, path string, width, height uint32) {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// The IHDR chunk follows the 8 bytes of the signature: length (4), type (4), width (4), height (4), 5 bytes, CRC.
	binary.BigEndian.PutUint32(b[16:20], width)
	binary.BigEndian.PutUint32(b[20:24], height)
	binary.BigEndian.PutUint32(b[29:33], crc32.ChecksumIEEE(b[12:29]))

	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestSave_TooManyPixels(t *testing.T) {
	src := writePNG(t, t.TempDir(), "bomb.png", 2, 2)
	withDimensions(t, src, 100_000, 100_000)

	if _, err := save(src, t.TempDir(), 10, 10, false, false, true); !errors.Is(err, ErrTooManyPixels) {
		t.Errorf("save() error = %v, want ErrTooManyPixels", err)
	}
}

func TestSave_InvalidHeader(t *testing.T) {
	src := filepath.Join(t.TempDir(), "text.png")
	if err := os.WriteFile(src, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := save(src, t.TempDir(), 10, 10, false, false, true); err == nil {
		t.Error("save() error = nil, want an error for an unreadable header")
	}
	if _, err := save(filepath.Join(t.TempDir(), "missing.png"), t.TempDir(), 10, 10, false, false, true); err == nil {
		t.Error("save() error = nil, want an error for a missing file")
	}
}
