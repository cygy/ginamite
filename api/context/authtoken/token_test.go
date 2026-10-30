package authtoken

import (
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"

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
