package request

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

// pngHeader is the signature of a PNG file, enough to be detected as an image.
var pngHeader = []byte("\x89PNG\r\n\x1a\n" + "0123456789012345678901234567890123456789")

func uploadContext(t *testing.T, field, filename string, content []byte) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("unable to create the form file: %v", err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatalf("unable to write the form file: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("unable to close the writer: %v", err)
	}

	c, recorder := apitest.NewContext("POST", "/", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	ctxLocale.Set(c, "", "en")

	return c, recorder
}

func TestGetUploadedFile(t *testing.T) {
	initLocalization(t)

	tests := []struct {
		name         string
		field        string
		content      []byte
		maxSize      int64
		checkIsImage bool
		wantFile     bool
		wantStatus   int
	}{
		{"valid image", "file", pngHeader, 1024, true, true, http.StatusOK},
		{"valid file without image check", "file", []byte("text"), 1024, false, true, http.StatusOK},
		{"missing field", "other", pngHeader, 1024, true, false, http.StatusInternalServerError},
		{"file too large", "file", pngHeader, 10, true, false, http.StatusPreconditionFailed},
		{"not an image", "file", []byte("this is only text, not an image"), 1024, true, false, http.StatusPreconditionFailed},
		{"empty file", "file", nil, 1024, true, false, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := uploadContext(t, tt.field, "image.png", tt.content)
			file := GetUploadedFile(c, "file", tt.maxSize, tt.checkIsImage, "test", "msg")
			c.Writer.WriteHeaderNow()

			if (file != nil) != tt.wantFile {
				t.Errorf("file returned = %v, want %v", file != nil, tt.wantFile)
			}
			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}

func TestSaveUploadedFile(t *testing.T) {
	initLocalization(t)

	t.Run("saved", func(t *testing.T) {
		c, _ := uploadContext(t, "file", "image.png", pngHeader)
		file := GetUploadedFile(c, "file", 1024, true, "test", "msg")
		destination := filepath.Join(t.TempDir(), "saved.png")

		if !SaveUploadedFile(c, file, destination, "test", "msg") {
			t.Fatal("SaveUploadedFile() = false, want true")
		}
		if content, err := os.ReadFile(destination); err != nil || !bytes.Equal(content, pngHeader) {
			t.Errorf("saved content = %q, %v, want the uploaded content", content, err)
		}
	})

	t.Run("unable to save", func(t *testing.T) {
		c, recorder := uploadContext(t, "file", "image.png", pngHeader)
		file := GetUploadedFile(c, "file", 1024, true, "test", "msg")
		// The parent of the destination is a regular file, so the destination can not be created.
		parent := filepath.Join(t.TempDir(), "file")
		if err := os.WriteFile(parent, []byte("x"), 0o600); err != nil {
			t.Fatalf("unable to write the file: %v", err)
		}
		destination := filepath.Join(parent, "saved.png")

		if SaveUploadedFile(c, file, destination, "test", "msg") {
			t.Fatal("SaveUploadedFile() = true, want false")
		}
		c.Writer.WriteHeaderNow()
		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}
	})
}
