package request

import (
	"net/http"
	"strings"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxkey"

	"github.com/gin-gonic/gin"
)

func TestUnescape(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"no backslash", "abc", "abc"},
		{"double backslash", `a\\b`, `a\b`},
		{"single backslash", `a\b`, `a\b`},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unescape(tt.in); got != tt.want {
				t.Errorf("Unescape(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDecodeBody(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	initLocalization(t)

	tests := []struct {
		name       string
		body       string
		limit      int64
		want       string
		wantOK     bool
		wantStatus int
	}{
		{"valid json", `{"name":"john"}`, 0, "john", true, http.StatusOK},
		{"empty body", ``, 0, "", true, http.StatusOK},
		{"invalid json", `{"name":`, 0, "", false, http.StatusBadRequest},
		{"wrong type", `{"name":1}`, 0, "", false, http.StatusBadRequest},
		{"over the limit", `{"name":"john"}`, 5, "", false, http.StatusRequestEntityTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := apitest.NewContext("POST", "/", strings.NewReader(tt.body))
			ctxLocale.Set(c, "", "en")
			if tt.limit > 0 {
				c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, tt.limit)
			}
			p := &payload{}

			if ok := DecodeBody(c, p); ok != tt.wantOK {
				t.Fatalf("DecodeBody() = %v, want %v", ok, tt.wantOK)
			}
			c.Writer.WriteHeaderNow()

			if p.Name != tt.want {
				t.Errorf("Name = %q, want %q", p.Name, tt.want)
			}
			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetOffsetAndLimit(t *testing.T) {
	initLocalization(t)

	tests := []struct {
		name       string
		query      string
		wantOK     bool
		wantOffset int
		wantLimit  int
	}{
		{"defaults", "", true, 0, 10},
		{"values", "?offset=5&limit=20", true, 5, 20},
		{"invalid values fall back to the defaults", "?offset=x&limit=y", true, 0, 10},
		{"negative offset", "?offset=-1", false, 0, 0},
		{"negative limit", "?limit=-1", false, 0, 0},
		{"limit above the maximum", "?limit=51", false, 0, 0},
		{"limit equal to the maximum", "?limit=50", true, 0, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := apitest.NewContext("GET", "/"+tt.query, nil)
			ctxLocale.Set(c, "", "en")

			ok, offset, limit := GetOffsetAndLimit(c, 0, 10, 50)
			if ok != tt.wantOK || offset != tt.wantOffset || limit != tt.wantLimit {
				t.Errorf("GetOffsetAndLimit() = (%v, %d, %d), want (%v, %d, %d)", ok, offset, limit, tt.wantOK, tt.wantOffset, tt.wantLimit)
			}
			c.Writer.WriteHeaderNow()
			if !tt.wantOK && recorder.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestGetOffsetAndGetLimit_NothingStored(t *testing.T) {
	c, _ := apitest.NewContext("GET", "/", nil)

	if GetOffset(c) != 0 || GetLimit(c) != 0 {
		t.Error("GetOffset and GetLimit must return 0 when nothing is stored")
	}
}

func TestGetLocaleValue(t *testing.T) {
	initLocalization(t)
	testConfig := &config.Configuration{SupportedLocales: []string{"en", "fr"}}

	tests := []struct {
		name       string
		target     string
		want       string
		wantOK     bool
		wantStatus int
	}{
		{"supported locale", "/FR", "fr", true, http.StatusOK},
		{"unsupported locale", "/es", "", false, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			var ok bool
			router := gin.New()
			router.Use(ctxkey.Config.Inject(testConfig), ctxLocaleMiddleware())
			router.GET("/:locale", func(c *gin.Context) {
				got, ok = GetLocaleValue(c, "locale", "msg")
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, tt.target, nil)
			recorder := apitest.Serve(router, req)

			if got != tt.want || ok != tt.wantOK {
				t.Errorf("GetLocaleValue() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.wantOK)
			}
			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetLocaleValue_Missing(t *testing.T) {
	initLocalization(t)

	c, recorder := apitest.NewContext("GET", "/", nil)
	ctxkey.Config.Set(c, config.NewConfiguration())
	ctxLocale.Set(c, "", "en")

	if got, ok := GetLocaleValue(c, "locale", "msg"); got != "" || ok {
		t.Errorf("GetLocaleValue() = (%q, %v), want (\"\", false)", got, ok)
	}
	c.Writer.WriteHeaderNow()
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
