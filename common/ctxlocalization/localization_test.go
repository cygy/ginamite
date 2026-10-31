package ctxlocalization

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestInject(t *testing.T) {
	engine := localization.NewEngine()

	var got *localization.Engine
	router := gin.New()
	router.Use(Inject(engine))
	router.GET("/", func(c *gin.Context) { got = Get(c) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got != engine {
		t.Errorf("Get() = %p, want %p", got, engine)
	}
}

func TestSet(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	engine := localization.NewEngine()

	Set(c, engine)

	if Get(c) != engine {
		t.Error("Get() must return the engine set")
	}
}

func TestGet_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	defer func() {
		if recover() == nil {
			t.Error("Get() must panic when there is no engine")
		}
	}()

	Get(c)
}
