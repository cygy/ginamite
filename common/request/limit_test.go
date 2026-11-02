package request

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLimitBodySize(t *testing.T) {
	tests := []struct {
		name       string
		maxSize    int64
		body       string
		chunked    bool
		wantStatus int
		wantRead   string
		wantErr    bool
	}{
		{"under the limit", 10, "hello", false, http.StatusOK, "hello", false},
		{"announced over the limit", 3, "hello", false, http.StatusRequestEntityTooLarge, "", false},
		{"chunked over the limit", 3, "hello", true, http.StatusOK, "hel", true},
		{"no limit", 0, "hello", false, http.StatusOK, "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var read string
			var readErr error

			engine := gin.New()
			engine.Use(LimitBodySize(tt.maxSize))
			engine.POST("/", func(c *gin.Context) {
				b, err := io.ReadAll(c.Request.Body)
				read, readErr = string(b), err
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(tt.body))
			if tt.chunked {
				req.ContentLength = -1
			}
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if read != tt.wantRead {
				t.Errorf("read %q, want %q", read, tt.wantRead)
			}
			var maxErr *http.MaxBytesError
			if errors.As(readErr, &maxErr) != tt.wantErr {
				t.Errorf("read error = %v, want a limit error: %v", readErr, tt.wantErr)
			}
		})
	}
}

func TestLimitBodySize_NoBody(t *testing.T) {
	engine := gin.New()
	engine.Use(LimitBodySize(1))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.Body = nil
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
