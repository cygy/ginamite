package user

import (
	"net/http"
	"testing"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	"github.com/cygy/ginamite/api/context/shared"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

func TestGet(t *testing.T) {
	c, _ := apitest.NewContext("GET", "/", nil)

	if got := Get(c); got != nil {
		t.Errorf("Get() = %v, want nil when no user is stored", got)
	}

	c.Set(shared.UserContextKey, "john")
	if got := Get(c); got != "john" {
		t.Errorf("Get() = %v, want john", got)
	}
}

func TestGetID(t *testing.T) {
	const secret = "secret"

	c, _ := apitest.NewContext("GET", "/", nil)
	if got := GetID(c); got != "" {
		t.Errorf("GetID() = %q, want an empty string without token", got)
	}

	router := gin.New()
	router.Use(ctxAuthToken.Parse(secret, nil, true))
	router.GET("/", func(c *gin.Context) { c.String(http.StatusOK, GetID(c)) })

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+apitest.SignToken(t, secret, "user-1", "token-1", false))

	if got := apitest.Serve(router, req).Body.String(); got != "user-1" {
		t.Errorf("GetID() = %q, want user-1", got)
	}
}
