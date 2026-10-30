package response

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/html"

	"github.com/gin-gonic/gin"
)

func TestErrorPages(t *testing.T) {
	tests := []struct {
		name    string
		handler gin.HandlerFunc
		status  int
	}{
		{"NotFound", NotFound, http.StatusNotFound},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized},
		{"InvalidAuthentication", InvalidAuthentication, http.StatusUnauthorized},
		{"ExpiredAuthentication", ExpiredAuthentication, http.StatusUnauthorized},
		{"RevokedAuthentication", RevokedAuthentication, http.StatusUnauthorized},
		{"InternalServerError", InternalServerError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, "en-us", nil, tt.handler)

			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			if rec.Body.Len() == 0 {
				t.Error("the page is empty")
			}
		})
	}
}

func TestInternalServerError_DisplaysTheRequestID(t *testing.T) {
	tests := []struct {
		locale string
		want   string
	}{
		{"en-us", "<html>|||Error request-1</html>"},
		{"fr-fr", "<html>|||Erreur request-1</html>"},
	}

	for _, tt := range tests {
		rec := serve(t, tt.locale, map[string]string{"X-Request-ID": "request-1"}, InternalServerError)

		if rec.Body.String() != tt.want {
			t.Errorf("%s: body = %q, want %q", tt.locale, rec.Body.String(), tt.want)
		}
	}
}

func TestErrorPages_WithoutPage(t *testing.T) {
	tests := []struct {
		name    string
		handler gin.HandlerFunc
		status  int
	}{
		{"NotFound", NotFound, http.StatusNotFound},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized},
		{"InvalidAuthentication", InvalidAuthentication, http.StatusUnauthorized},
		{"InternalServerError", InternalServerError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, "de-de", nil, tt.handler)

			if rec.Code != tt.status || rec.Body.Len() != 0 {
				t.Errorf("status = %d, body = %q, want only the status %d", rec.Code, rec.Body.String(), tt.status)
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	rec := serve(t, "en-us", nil, func(c *gin.Context) {
		page := html.Main.PageByKey(c, "index")

		APIError(c, page, response.Error{Message: "It failed"})
	})

	if want := "<html>Title|It failed||</html>"; rec.Code != http.StatusOK || rec.Body.String() != want {
		t.Errorf("status = %d, body = %q, want %q", rec.Code, rec.Body.String(), want)
	}
}
