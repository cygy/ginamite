package shorthand

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/internal/pagetest"

	"github.com/gin-gonic/gin"
)

// run initializes the pages, then sends a GET request on the locale with the query to the handler, which runs after
// the middleware saving the page key.
func run(t *testing.T, pageKey, query string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	engine := pagetest.Setup(t)
	engine.GET("/validate", context.SaveLocale("en-us"), context.SavePageKey(pageKey), handler)

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/validate?"+query, nil))

	return rec
}

func TestNewValidation(t *testing.T) {
	var v *Validation
	run(t, "index", "", func(c *gin.Context) { v = NewValidation(c) })

	if v == nil || v.PageKey != "index" || v.APICall != nil {
		t.Errorf("Validation = %+v", v)
	}
}

func TestValidation_Go_Success(t *testing.T) {
	var gotID, gotKey string
	rec := run(t, "index", "id=u1&key=k1", func(c *gin.Context) {
		v := NewValidation(c)
		v.APICall = func(c *gin.Context, id, key string) (string, api.Result, bool) {
			gotID, gotKey = id, key
			return "Validated", api.Result{StatusCode: http.StatusOK}, true
		}
		v.Go(c)
	})

	if gotID != "u1" || gotKey != "k1" {
		t.Errorf("id = %q, key = %q", gotID, gotKey)
	}
	if want := "<html>Title||Validated|</html>"; rec.Code != http.StatusOK || rec.Body.String() != want {
		t.Errorf("status = %d, body = %q, want %q", rec.Code, rec.Body.String(), want)
	}
}

func TestValidation_Go_Hooks(t *testing.T) {
	var calls []string
	rec := run(t, "index", "", func(c *gin.Context) {
		v := NewValidation(c)
		v.PageConfiguration = func(page *html.Page) {
			calls = append(calls, "configuration")
			page.SetVar("ID", "configured")
		}
		v.APICall = func(c *gin.Context, id, key string) (string, api.Result, bool) {
			calls = append(calls, "call")
			return "Done", api.Result{}, true
		}
		v.BeforeRendering = func(c *gin.Context, page *html.Page, message string, result api.Result, ok bool) {
			calls = append(calls, "before:"+message)
			page.SetVar("ID", page.Var("ID").(string)+"+before")
		}
		v.RenderOverride = func(c *gin.Context, page *html.Page, message string, result api.Result, ok bool) {
			calls = append(calls, "override")
			c.String(http.StatusAccepted, "overridden %v", page.Var("ID"))
		}
		v.Go(c)
	})

	if want := []string{"configuration", "call", "before:Done", "override"}; !reflect.DeepEqual(calls, want) {
		t.Errorf("calls = %v, want %v", calls, want)
	}
	if rec.Code != http.StatusAccepted || rec.Body.String() != "overridden configured+before" {
		t.Errorf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
}

func TestValidation_Go_Failure(t *testing.T) {
	tests := []struct {
		name       string
		result     api.Result
		wantStatus int
		wantBody   string
	}{
		{"API error", api.Result{StatusCode: http.StatusBadRequest, APIError: response.Error{Code: 1, Domain: "test", Message: "Invalid key"}}, http.StatusOK, "<html>Title|Invalid key||</html>"},
		{"not found", api.Result{StatusCode: http.StatusNotFound}, http.StatusNotFound, "<html>|||</html>"},
		{"unauthorized", api.Result{StatusCode: http.StatusUnauthorized}, http.StatusUnauthorized, "<html>|||</html>"},
		{"another status", api.Result{StatusCode: http.StatusTeapot}, http.StatusInternalServerError, "<html>|||Error </html>"},
		{"handled", api.Result{StatusCode: http.StatusBadRequest, Handled: true}, http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rendered bool
			rec := run(t, "index", "", func(c *gin.Context) {
				v := NewValidation(c)
				v.APICall = func(c *gin.Context, id, key string) (string, api.Result, bool) { return "", tt.result, false }
				v.BeforeRendering = func(*gin.Context, *html.Page, string, api.Result, bool) { rendered = true }
				v.Go(c)
			})

			if rec.Code != tt.wantStatus || rec.Body.String() != tt.wantBody {
				t.Errorf("status = %d, body = %q, want %d, %q", rec.Code, rec.Body.String(), tt.wantStatus, tt.wantBody)
			}
			if rendered {
				t.Error("BeforeRendering must not be called when the call fails")
			}
		})
	}
}

func TestValidation_Go_UnknownPage(t *testing.T) {
	called := false
	rec := run(t, "unknown", "", func(c *gin.Context) {
		v := NewValidation(c)
		v.APICall = func(c *gin.Context, id, key string) (string, api.Result, bool) {
			called = true
			return "", api.Result{}, true
		}
		v.Go(c)
	})

	if called || rec.Code != http.StatusNotFound {
		t.Errorf("called = %v, status = %d", called, rec.Code)
	}
}
