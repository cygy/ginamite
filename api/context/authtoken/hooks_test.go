package authtoken

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

func testHooks() authentication.Hooks {
	return authentication.Hooks{
		BuildTokenFromID: func(context.Context, string) (*authentication.Token, error) { return nil, nil },
	}
}

func TestInjectHooks(t *testing.T) {
	var got authentication.Hooks
	router := gin.New()
	router.Use(InjectHooks(testHooks()))
	router.GET("/", func(c *gin.Context) { got = HooksOf(c) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got.BuildTokenFromID == nil {
		t.Error("HooksOf() must return the hooks of the middleware")
	}
}

func TestSetHooks(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	SetHooks(c, testHooks())

	if HooksOf(c).BuildTokenFromID == nil {
		t.Error("HooksOf() must return the hooks set")
	}
}

func TestHooksOf_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	defer func() {
		if recover() == nil {
			t.Error("HooksOf() must panic when there are no hooks")
		}
	}()

	HooksOf(c)
}

func TestInjectRevoked(t *testing.T) {
	var revoked, other bool
	router := gin.New()
	router.Use(InjectRevoked(func(tokenID string) bool { return tokenID == "revoked" }))
	router.GET("/", func(c *gin.Context) { revoked, other = isRevoked(c, "revoked"), isRevoked(c, "other") })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !revoked || other {
		t.Errorf("isRevoked() = %v, %v, want true, false", revoked, other)
	}
}

func TestSetRevoked(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	if isRevoked(c, "revoked") {
		t.Error("without function, no token is revoked")
	}

	SetRevoked(c, func(string) bool { return true })

	if !isRevoked(c, "revoked") {
		t.Error("isRevoked() must use the function set")
	}
}
