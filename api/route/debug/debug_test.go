package debug

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"

	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"
)

func TestGetExpiredToken(t *testing.T) {
	testConfig := &config.Configuration{}
	testConfig.JWT.Secret = "secret"
	testConfig.JWT.SigningMethod = "HS256"

	var hooks authentication.Hooks

	t.Run("token functions undefined", func(t *testing.T) {
		c, recorder := apitest.NewContext("GET", "/", nil)
		ctxconfig.Set(c, testConfig)
		ctxAuthToken.SetHooks(c, hooks)
		GetExpiredToken(c)
		c.Writer.WriteHeaderNow()

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		hooks.BuildTokenFromID = func(context.Context, string) (*authentication.Token, error) {
			return &authentication.Token{}, nil
		}
		hooks.ExtraPropertiesForTokenWithID = func(string) (map[string]string, error) {
			return map[string]string{}, nil
		}

		c, recorder := apitest.NewContext("GET", "/", nil)
		ctxconfig.Set(c, testConfig)
		ctxAuthToken.SetHooks(c, hooks)
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
		if token, err := authentication.ParseToken(body.AuthToken, "secret"); token != nil || err == nil {
			t.Errorf("ParseToken() = %v, %v, want an expiration error", token, err)
		}
	})
}
