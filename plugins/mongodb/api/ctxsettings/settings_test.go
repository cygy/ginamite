package ctxsettings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

func TestInjectSettings(t *testing.T) {
	settings := model.DefaultSettings()
	settings.UserMaxDescriptionLength = 12

	var got model.Settings
	router := gin.New()
	router.Use(InjectSettings(settings))
	router.GET("/", func(c *gin.Context) { got = GetSettings(c) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got.UserMaxDescriptionLength != 12 {
		t.Errorf("settings = %+v", got)
	}
}

func TestSetSettings(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	SetSettings(c, model.Settings{UserImagePath: "/custom"})

	if GetSettings(c).UserImagePath != "/custom" {
		t.Error("GetSettings() must return the settings set")
	}
}

func TestGetSettings_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	defer func() {
		if recover() == nil {
			t.Error("GetSettings() must panic when there are no settings")
		}
	}()

	GetSettings(c)
}
