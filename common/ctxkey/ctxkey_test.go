package ctxkey

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestKey_Inject(t *testing.T) {
	configuration := config.NewConfiguration()

	var got *config.Configuration
	router := gin.New()
	router.Use(Config.Inject(configuration))
	router.GET("/", func(c *gin.Context) { got = Config.Get(c) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got != configuration {
		t.Errorf("Get() = %p, want %p", got, configuration)
	}
}

func TestKey_Set(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	configuration := config.NewConfiguration()

	Config.Set(c, configuration)

	if Config.Get(c) != configuration {
		t.Error("Get() must return the service set")
	}
}

func TestKey_KeysAreIndependent(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	first, second := New[string]("first"), New[string]("second")

	first.Set(c, "a")
	second.Set(c, "b")

	if first.Get(c) != "a" || second.Get(c) != "b" {
		t.Errorf("Get() = %q and %q, want a and b", first.Get(c), second.Get(c))
	}
}

func TestKey_Get_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	defer func() {
		if got := recover(); got != "ctxkey: no queue in the context, add the middleware Inject" {
			t.Errorf("recover() = %v", got)
		}
	}()

	Queue.Get(c)
}
