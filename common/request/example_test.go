package request_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cygy/ginamite/common/request"

	"github.com/gin-gonic/gin"
)

// The bodies larger than the limit are rejected before a handler reads them.
func ExampleLimitBodySize() {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(request.LimitBodySize(16))
	engine.POST("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for _, body := range []string{`{"a":1}`, `{"name":"a very long name"}`} {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))
		fmt.Println(recorder.Code)
	}
	// Output:
	// 204
	// 413
}

// The cross-origin requests which change something are rejected, except from the trusted origins.
func ExampleCrossOriginProtection() {
	gin.SetMode(gin.TestMode)
	protection, err := request.CrossOriginProtection("https://www.example.com")
	if err != nil {
		panic(err)
	}

	engine := gin.New()
	engine.Use(protection)
	engine.POST("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for _, origin := range []string{"https://www.example.com", "https://evil.example"} {
		req := httptest.NewRequest(http.MethodPost, "http://web.internal/", nil)
		req.Header.Set("Origin", origin)
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, req)
		fmt.Println(origin, recorder.Code)
	}
	// Output:
	// https://www.example.com 204
	// https://evil.example 403
}
