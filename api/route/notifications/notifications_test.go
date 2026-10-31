package notifications

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/ctxlocalization"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

func TestGetNotifications(t *testing.T) {
	apitest.InitLocalization(t, map[string]string{
		"notification.login.name":        "Login",
		"notification.login.description": "A login occurs",
	})

	router := gin.New()
	router.Use(ctxlocalization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
	router.GET("/", GetNotifications([]string{"login"}))

	// The second call is served from the cache.
	for _, call := range []string{"first", "second"} {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		recorder := apitest.Serve(router, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s call: status = %d, want %d", call, recorder.Code, http.StatusOK)
		}

		var body struct {
			Notifications []localizedNotification `json:"notifications"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s call: invalid body: %v", call, err)
		}
		want := localizedNotification{Type: "login", Name: "Login", Description: "A login occurs"}
		if len(body.Notifications) != 1 || body.Notifications[0] != want {
			t.Errorf("%s call: notifications = %+v, want [%+v]", call, body.Notifications, want)
		}
	}
}
