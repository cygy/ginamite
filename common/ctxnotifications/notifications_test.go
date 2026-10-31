package ctxnotifications

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/common/notifications"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestInject(t *testing.T) {
	registry := notifications.NewRegistry()

	var got *notifications.Registry
	router := gin.New()
	router.Use(Inject(registry))
	router.GET("/", func(c *gin.Context) { got = Get(c) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got != registry {
		t.Errorf("Get() = %p, want %p", got, registry)
	}
}

func TestSet(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	registry := notifications.NewRegistry()

	Set(c, registry)

	if Get(c) != registry {
		t.Error("Get() must return the registry set")
	}
}

func TestGet_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	defer func() {
		if recover() == nil {
			t.Error("Get() must panic when there is no registry")
		}
	}()

	Get(c)
}
