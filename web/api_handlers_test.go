package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/internal/pagetest"

	"github.com/gin-gonic/gin"
)

// serve initializes the pages then sends a GET request to the handler, and returns the response.
func serve(t *testing.T, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	engine := pagetest.Setup(t)
	engine.GET("/test", context.SaveLocale("en-us"), handler)

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test", nil))

	return rec
}

func TestHandleAPICallToPage(t *testing.T) {
	apiError := response.Error{Code: 1, Domain: "test", Message: "It failed"}
	tests := []struct {
		name        string
		result      api.Result
		ok          bool
		wantHandled bool
		wantStatus  int
		wantBody    string
	}{
		{"success", api.Result{StatusCode: http.StatusOK}, true, false, http.StatusOK, ""},
		{"already handled", api.Result{StatusCode: http.StatusBadRequest, Handled: true}, false, true, http.StatusOK, ""},
		{"API error", api.Result{StatusCode: http.StatusBadRequest, APIError: apiError}, false, true, http.StatusOK, "<html>Title|It failed||</html>"},
		{"not found", api.Result{StatusCode: http.StatusNotFound}, false, true, http.StatusNotFound, "<html>|||</html>"},
		{"unauthorized", api.Result{StatusCode: http.StatusUnauthorized}, false, true, http.StatusUnauthorized, "<html>|||</html>"},
		{"internal server error", api.Result{StatusCode: http.StatusInternalServerError}, false, true, http.StatusInternalServerError, "<html>|||Error </html>"},
		{"another status", api.Result{StatusCode: http.StatusForbidden}, false, true, http.StatusInternalServerError, "<html>|||Error </html>"},
		{"no response", api.Result{RequestError: errors.New("timeout")}, false, true, http.StatusInternalServerError, "<html>|||Error </html>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handled bool
			rec := serve(t, func(c *gin.Context) {
				handled = HandleAPICallToPage(c, html.Main.PageByKey(c, "index"), tt.result, tt.ok)
			})

			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
			if rec.Code != tt.wantStatus || rec.Body.String() != tt.wantBody {
				t.Errorf("status = %d, body = %q, want %d, %q", rec.Code, rec.Body.String(), tt.wantStatus, tt.wantBody)
			}
		})
	}
}

func TestHandleAPICallToPartialPage(t *testing.T) {
	tests := []struct {
		name        string
		result      api.Result
		ok          bool
		wantHandled bool
		wantStatus  int
		wantBody    string
	}{
		{"success", api.Result{StatusCode: http.StatusOK}, true, false, http.StatusOK, ""},
		{"already handled", api.Result{StatusCode: http.StatusBadRequest, Handled: true}, false, true, http.StatusOK, ""},
		{"API error", api.Result{StatusCode: http.StatusBadRequest, APIError: response.Error{Code: 1, Domain: "test", Message: "It failed"}}, false, true, http.StatusBadRequest, `{"error":{"code":1,"domain":"test","message":"It failed"}}`},
		{"no response", api.Result{RequestError: errors.New("timeout")}, false, true, http.StatusInternalServerError, `{"error":{"code":0,"domain":""}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handled bool
			rec := serve(t, func(c *gin.Context) {
				handled = HandleAPICallToPartialPage(c, tt.result, tt.ok)
			})

			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
			if rec.Code != tt.wantStatus || rec.Body.String() != tt.wantBody {
				t.Errorf("status = %d, body = %q, want %d, %q", rec.Code, rec.Body.String(), tt.wantStatus, tt.wantBody)
			}
		})
	}
}
