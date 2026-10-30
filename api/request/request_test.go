package request

import (
	"net/http"
	"strings"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

func initLocalization(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{
		"msg": "message", "reason": "reason", "recovery": "recovery",
		"error.request.parameter.offset.negative.message":  "m",
		"error.request.parameter.offset.negative.reason":   "r",
		"error.request.parameter.offset.negative.recovery": "r",
		"error.request.parameter.limit.negative.message":   "m",
		"error.request.parameter.limit.negative.reason":    "r",
		"error.request.parameter.limit.negative.recovery":  "r",
		"error.request.parameter.limit.max.message":        "m",
		"error.request.parameter.limit.max.reason":         "max {{.LimitMax}}",
		"error.request.parameter.limit.max.recovery":       "max {{.LimitMax}}",
	})
}

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

	tests := []struct {
		name string
		body string
		want string
	}{
		{"valid json", `{"name":"john"}`, "john"},
		{"invalid json is ignored", `{"name":`, ""},
		{"empty body", ``, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := apitest.NewContext("POST", "/", strings.NewReader(tt.body))
			p := &payload{}
			DecodeBody(c, p)

			if p.Name != tt.want {
				t.Errorf("Name = %q, want %q", p.Name, tt.want)
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

func TestVerifyLimitAndOffset(t *testing.T) {
	initLocalization(t)

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantOffset int64
		wantLimit  int64
	}{
		{"defaults", "", http.StatusOK, 0, 10},
		{"values", "?offset=3&limit=4", http.StatusOK, 3, 4},
		{"invalid", "?offset=-3", http.StatusBadRequest, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var offset, limit int64
			router := gin.New()
			router.Use(ctxLocaleMiddleware(), VerifyLimitAndOffset(0, 10, 50))
			router.GET("/", func(c *gin.Context) {
				offset, limit = GetOffset(c), GetLimit(c)
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, "/"+tt.query, nil)
			recorder := apitest.Serve(router, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if offset != tt.wantOffset || limit != tt.wantLimit {
				t.Errorf("offset, limit = %d, %d, want %d, %d", offset, limit, tt.wantOffset, tt.wantLimit)
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
	config.Main = &config.Configuration{SupportedLocales: []string{"en", "fr"}}

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
			router.Use(ctxLocaleMiddleware())
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
	ctxLocale.Set(c, "", "en")

	if got, ok := GetLocaleValue(c, "locale", "msg"); got != "" || ok {
		t.Errorf("GetLocaleValue() = (%q, %v), want (\"\", false)", got, ok)
	}
	c.Writer.WriteHeaderNow()
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestVerifyLocaleParameter(t *testing.T) {
	initLocalization(t)
	config.Main = &config.Configuration{SupportedLocales: []string{"en", "fr"}}

	var stored any
	router := gin.New()
	router.Use(ctxLocaleMiddleware(), VerifyLocaleParameter(Config{Name: "locale", StoreKey: "stored", MessageKey: "msg"}))
	router.GET("/:locale", func(c *gin.Context) {
		stored, _ = c.Get("stored")
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/FR", nil)
	if recorder := apitest.Serve(router, req); recorder.Code != http.StatusOK || stored != "fr" {
		t.Errorf("status = %d, stored = %v, want 200, fr", recorder.Code, stored)
	}

	req, _ = http.NewRequest(http.MethodGet, "/es", nil)
	if recorder := apitest.Serve(router, req); recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestVerifyURLParameter(t *testing.T) {
	initLocalization(t)

	const validID = "507f1f77bcf86cd799439011"

	tests := []struct {
		name       string
		cfg        Config
		target     string
		wantStatus int
		wantStored any
	}{
		{"valid object ID", Config{Name: "id", StoreKey: "stored"}, "/" + validID, http.StatusOK, validID},
		{"invalid object ID", Config{Name: "id", StoreKey: "stored", InvalidObjectIDReasonKey: "reason"}, "/xyz", http.StatusBadRequest, nil},
		{"invalid object ID accepted when skipped", Config{Name: "id", StoreKey: "stored", SkipVerifyingInvalidObjectID: true}, "/xyz", http.StatusOK, "xyz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stored any
			router := gin.New()
			router.Use(ctxLocaleMiddleware(), VerifyURLParameter(tt.cfg))
			router.GET("/:id", func(c *gin.Context) {
				stored, _ = c.Get("stored")
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, tt.target, nil)
			recorder := apitest.Serve(router, req)

			if recorder.Code != tt.wantStatus || stored != tt.wantStored {
				t.Errorf("status = %d, stored = %v, want %d, %v", recorder.Code, stored, tt.wantStatus, tt.wantStored)
			}
		})
	}
}

func TestVerifyURLParameter_Missing(t *testing.T) {
	initLocalization(t)

	c, recorder := apitest.NewContext("GET", "/", nil)
	ctxLocale.Set(c, "", "en")
	VerifyURLParameter(Config{Name: "id", StoreKey: "stored"})(c)
	c.Writer.WriteHeaderNow()

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

// ctxLocaleMiddleware stores the "en" locale in the context.
func ctxLocaleMiddleware() gin.HandlerFunc {
	return ctxLocale.Inject("en", []string{"en", "fr"})
}
