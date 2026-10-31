package request

import (
	"net/http"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

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
	setIDValidator(t, func(id string) bool { return id == validID })

	tests := []struct {
		name       string
		cfg        Config
		target     string
		wantStatus int
		wantStored any
	}{
		{"valid ID", Config{Name: "id", StoreKey: "stored"}, "/" + validID, http.StatusOK, validID},
		{"invalid ID", Config{Name: "id", StoreKey: "stored", InvalidIDReasonKey: "reason"}, "/xyz", http.StatusBadRequest, nil},
		{"invalid ID accepted when skipped", Config{Name: "id", StoreKey: "stored", SkipVerifyingInvalidID: true}, "/xyz", http.StatusOK, "xyz"},
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
