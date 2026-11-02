package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsValidID(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	for _, id := range []string{"", "xyz", "507f1f77bcf86cd799439011"} {
		if !IsValidID(c, id) {
			t.Errorf("IsValidID(%q) = false, want true without validator", id)
		}
	}

	SetIDValidator(c, func(id string) bool { return id == "ok" })

	if !IsValidID(c, "ok") || IsValidID(c, "xyz") {
		t.Error("IsValidID() must use the validator of the context")
	}
}

func TestInjectIDValidator(t *testing.T) {
	var valid, invalid bool
	router := gin.New()
	router.Use(InjectIDValidator(func(id string) bool { return id == "ok" }))
	router.GET("/", func(c *gin.Context) { valid, invalid = IsValidID(c, "ok"), IsValidID(c, "xyz") })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !valid || invalid {
		t.Errorf("IsValidID() = %v, %v, want true, false", valid, invalid)
	}
}
