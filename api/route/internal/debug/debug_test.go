package debug

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxkey"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestGetExpiredToken(t *testing.T) {
	testConfig := &config.Configuration{}
	testConfig.JWT.Secret = "secret"
	testConfig.JWT.SigningMethod = "HS256"

	c, recorder := apitest.NewContext("GET", "/", nil)
	ctxkey.Config.Set(c, testConfig)
	GetExpiredToken(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body struct {
		AuthToken string `json:"auth_token"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid body: %v", err)
	}
	if _, err := authentication.ParseToken(body.AuthToken, "secret"); !errors.Is(err, jwt.ErrTokenExpired) {
		t.Errorf("ParseToken() error = %v, want an expired token signed by the API", err)
	}
}
