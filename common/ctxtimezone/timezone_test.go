package ctxtimezone

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestInject(t *testing.T) {
	timezones := timezone.NewConfiguration()

	var got *timezone.Configuration
	router := gin.New()
	router.Use(Inject(timezones))
	router.GET("/", func(c *gin.Context) { got = Get(c) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got != timezones {
		t.Errorf("Get() = %p, want %p", got, timezones)
	}
}

func TestSet(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	timezones := timezone.NewConfiguration()

	Set(c, timezones)

	if Get(c) != timezones {
		t.Error("Get() must return the timezones set")
	}
}

func TestGet_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	defer func() {
		if recover() == nil {
			t.Error("Get() must panic when there are no timezones")
		}
	}()

	Get(c)
}
