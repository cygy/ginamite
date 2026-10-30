package authtoken

import (
	"net/http"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

func TestGetters_NothingStored(t *testing.T) {
	c, _ := apitest.NewContext("GET", "/", nil)

	if GetSigned(c) != "" || GetParsed(c) != nil || GetID(c) != "" {
		t.Error("getters must return zero values when no token is stored")
	}
}

func TestGetters_TokenStored(t *testing.T) {
	c, _ := apitest.NewContext("GET", "/", nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": "token-1"})
	c.Set(authTokenContextKey, "signed")
	c.Set(authTokenParsedContextKey, token)

	if got := GetSigned(c); got != "signed" {
		t.Errorf("GetSigned() = %q, want signed", got)
	}
	if got := GetParsed(c); got != token {
		t.Errorf("GetParsed() = %v, want the stored token", got)
	}
	if got := GetID(c); got != "token-1" {
		t.Errorf("GetID() = %q, want token-1", got)
	}
}

func TestVerify(t *testing.T) {
	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	tests := []struct {
		name       string
		verify     IsValidFunc
		wantStatus int
	}{
		{"no function", nil, http.StatusOK},
		{"valid token", func(*gin.Context, string) (bool, error) { return true, nil }, http.StatusOK},
		{"invalid token", func(*gin.Context, string) (bool, error) { return false, nil }, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := apitest.NewContext("GET", "/", nil)
			ctxLocale.Set(c, "", "en")
			Verify(tt.verify)(c)
			c.Writer.WriteHeaderNow()

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
