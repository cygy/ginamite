package pagetest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

func TestSetup(t *testing.T) {
	originalConfig, originalHTML := webtest.Config, html.Main

	t.Run("initialized", func(t *testing.T) {
		engine := Setup(t)

		if webtest.Config.Cookie.Name != CookieName {
			t.Errorf("cookie name = %q", webtest.Config.Cookie.Name)
		}
		if got := html.Main.URL("index", "fr-fr"); got != WebHost+"/fr-fr/accueil" {
			t.Errorf("URL() = %q", got)
		}

		engine.GET("/t", func(c *gin.Context) {
			c.Set("_locale", "en-us")
			html.Main.PageByKey(c, "index").RenderOk(c)
		})
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/t", nil))

		if want := "<html>Title|||</html>"; rec.Body.String() != want {
			t.Errorf("body = %q, want %q", rec.Body.String(), want)
		}
	})

	if webtest.Config != originalConfig || html.Main != originalHTML {
		t.Error("the global state is not restored")
	}
}
