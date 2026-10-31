package ctxqueue

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestInject(t *testing.T) {
	q := queue.New(nil)

	var got *queue.Queue
	router := gin.New()
	router.Use(Inject(q))
	router.GET("/", func(c *gin.Context) { got = Get(c) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got != q {
		t.Errorf("Get() = %p, want %p", got, q)
	}
}

func TestSet(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	q := queue.New(nil)

	Set(c, q)

	if Get(c) != q {
		t.Error("Get() must return the queue set")
	}
}

func TestGet_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	defer func() {
		if recover() == nil {
			t.Error("Get() must panic when there is no queue")
		}
	}()

	Get(c)
}
