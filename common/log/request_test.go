package log

import (
	"bytes"
	"compress/gzip"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// do sends a request through the logging middleware and returns the logged fields of the request.
func do(t *testing.T, logBody bool, handler gin.HandlerFunc, req *http.Request) (logrus.Fields, *httptest.ResponseRecorder, *http.Request) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	hook := useHook(t)

	var seen *http.Request
	router := gin.New()
	router.Use(InjectRequestLogger(logBody))
	router.Any("/*path", func(c *gin.Context) {
		seen = c.Request
		handler(c)
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	entry := hook.LastEntry()
	if entry == nil || entry.Message != "request" {
		t.Fatalf("the request is not logged: %+v", entry)
	}

	return entry.Data, rec, seen
}

func gzipped(t *testing.T, s string) []byte {
	t.Helper()

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

func TestInjectRequestLogger_Request(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login?x=1", strings.NewReader(`{"password":"secret","email":"john.doe@example.com"}`))
	req.Header.Set("X-User-IP", "1.2.3.4")

	var body string
	fields, rec, seen := do(t, false, func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		body = string(b)
		c.String(http.StatusCreated, "ok")
	}, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d", rec.Code)
	}
	if !strings.Contains(body, "secret") || !strings.Contains(body, "john.doe@") {
		t.Errorf("the handler must still read the original body, got %q", body)
	}
	if seen.Header.Get(RequestIDHeaderName) == "" {
		t.Error("a request ID must be generated")
	}

	if fields["IP"] != "1.2.3.4" {
		t.Errorf("IP = %v", fields["IP"])
	}

	logged := fields["request"].(logrus.Fields)
	if logged["method"] != http.MethodPost || logged["URI"] != "/login?x=1" {
		t.Errorf("request = %v", logged)
	}
	if got := logged["body"].(string); strings.Contains(got, "secret") || strings.Contains(got, "john.doe") || !strings.Contains(got, `"password":"***"`) {
		t.Errorf("body not obfuscated: %q", got)
	}

	response := fields["response"].(logrus.Fields)
	if response["code"] != http.StatusCreated {
		t.Errorf("response = %v", response)
	}
	if _, ok := response["body"]; ok {
		t.Error("the response body must not be logged")
	}
}

func TestInjectRequestLogger_KeepsRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeaderName, "abc")

	_, _, seen := do(t, false, func(c *gin.Context) {}, req)

	if got := seen.Header.Get(RequestIDHeaderName); got != "abc" {
		t.Errorf("request ID = %q, want abc", got)
	}
}

func TestInjectRequestLogger_Multipart(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("a", "b")
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	fields, _, _ := do(t, false, func(c *gin.Context) {}, req)

	if got := fields["request"].(logrus.Fields)["body"]; got != "not logged" {
		t.Errorf("body = %v, want %q", got, "not logged")
	}
}

func TestInjectRequestLogger_NoBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Body = nil

	fields, _, _ := do(t, false, func(c *gin.Context) {}, req)

	if got := fields["request"].(logrus.Fields)["body"]; got != "" {
		t.Errorf("body = %v, want empty", got)
	}
}

func TestInjectRequestLogger_Error(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"string", "boom", "boom"},
		{"not a string", 42, "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, _, _ := do(t, false, func(c *gin.Context) { c.Set(ErrorContextKey, tt.value) }, httptest.NewRequest(http.MethodGet, "/", nil))

			if fields["error"] != tt.want {
				t.Errorf("error = %v, want %q", fields["error"], tt.want)
			}
		})
	}
}

func TestInjectRequestLogger_ResponseBody(t *testing.T) {
	tests := []struct {
		name      string
		write     func(c *gin.Context)
		wantBody  string
		wantError string
	}{
		{
			"plain with email",
			func(c *gin.Context) { c.String(http.StatusOK, `{"email":"john.doe@example.com"}`) },
			`{"email":"***@example.com"}`, "",
		},
		{
			"gzipped",
			func(c *gin.Context) { c.Data(http.StatusOK, "application/json", gzipped(t, "hello john@x.io")) },
			"***@x.io", "",
		},
		{
			"corrupted gzip",
			func(c *gin.Context) { c.Data(http.StatusOK, "text/plain", []byte{31, 139, 1, 2, 3}) },
			"", "can not uncompress body",
		},
		{
			"truncated gzip",
			func(c *gin.Context) { c.Data(http.StatusOK, "text/plain", gzipped(t, strings.Repeat("abc", 100))[:20]) },
			"", "can not uncompress body",
		},
		{
			"too short to be gzipped",
			func(c *gin.Context) { c.String(http.StatusOK, "a") },
			"a", "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, rec, _ := do(t, true, tt.write, httptest.NewRequest(http.MethodGet, "/", nil))

			response := fields["response"].(logrus.Fields)
			if response["body"] != tt.wantBody {
				t.Errorf("body = %q, want %q", response["body"], tt.wantBody)
			}
			if got, _ := response["error"].(string); got != tt.wantError {
				t.Errorf("error = %q, want %q", got, tt.wantError)
			}
			if rec.Body.Len() == 0 {
				t.Error("the client must still receive the response")
			}
		})
	}
}
