package ctxconfig

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestInject(t *testing.T) {
	configuration := config.NewConfiguration()

	var got *config.Configuration
	router := gin.New()
	router.Use(Inject(configuration))
	router.GET("/", func(c *gin.Context) { got = Get(c) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got != configuration {
		t.Errorf("Get() = %p, want %p", got, configuration)
	}
}

func TestSet(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	configuration := config.NewConfiguration()

	Set(c, configuration)

	if Get(c) != configuration {
		t.Error("Get() must return the configuration set")
	}
}

func TestGet_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	defer func() {
		if recover() == nil {
			t.Error("Get() must panic when there is no configuration")
		}
	}()

	Get(c)
}
