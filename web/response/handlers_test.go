package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authErrors "github.com/cygy/ginamite/common/errors/auth"
	termsErrors "github.com/cygy/ginamite/common/errors/terms"
	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/internal/pagetest"

	"github.com/gin-gonic/gin"
)

// handle runs the handler on a request and returns whether it handled it, whether it aborted the request, and the
// response.
func handle(t *testing.T, handler api.ResponseHandlerFunc, statusCode int, requestError error, apiError response.Error) (handled, aborted bool, rec *httptest.ResponseRecorder) {
	t.Helper()

	rec = serve(t, "en-us", nil, func(c *gin.Context) {
		handled = handler(c, statusCode, requestError, apiError)
		aborted = c.IsAborted()
	})

	return handled, aborted, rec
}

func TestHandleInternalServerError(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		requestError error
		wantHandled  bool
	}{
		{"status 500", http.StatusInternalServerError, nil, true},
		{"request error", 0, errors.New("timeout"), true},
		{"status 200", http.StatusOK, nil, false},
		{"status 404", http.StatusNotFound, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled, aborted, rec := handle(t, HandleInternalServerError(), tt.statusCode, tt.requestError, response.Error{})
			status, body := rec.Code, rec.Body.String()

			if handled != tt.wantHandled || aborted != tt.wantHandled {
				t.Errorf("handled = %v, aborted = %v, want %v", handled, aborted, tt.wantHandled)
			}
			if tt.wantHandled && (status != http.StatusInternalServerError || body == "") {
				t.Errorf("status = %d, body = %q", status, body)
			}
			if !tt.wantHandled && body != "" {
				t.Errorf("body = %q, want empty", body)
			}
		})
	}
}

func TestHandleInvalidAuthenticationError(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		apiError        response.Error
		wantHandled     bool
		wantInvalidated bool
	}{
		{"invalid token", http.StatusUnauthorized, response.Error{Domain: authErrors.Domain, Code: authErrors.InvalidAuthorizationToken}, true, true},
		{"expired token", http.StatusUnauthorized, response.Error{Domain: authErrors.Domain, Code: authErrors.ExpiredAuthorizationToken}, true, true},
		{"revoked token", http.StatusUnauthorized, response.Error{Domain: authErrors.Domain, Code: authErrors.RevokedAuthorizationToken}, true, true},
		{"insufficient rights", http.StatusUnauthorized, response.Error{Domain: authErrors.Domain, Code: authErrors.InsufficientRights}, true, false},
		{"another authentication error", http.StatusUnauthorized, response.Error{Domain: authErrors.Domain, Code: 999}, false, false},
		{"another domain", http.StatusUnauthorized, response.Error{Domain: "other", Code: authErrors.InvalidAuthorizationToken}, false, false},
		{"another status", http.StatusForbidden, response.Error{Domain: authErrors.Domain, Code: authErrors.InvalidAuthorizationToken}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled, aborted, rec := handle(t, HandleInvalidAuthenticationError(), tt.statusCode, nil, tt.apiError)
			header, status, body := rec.Header(), rec.Code, rec.Body.String()

			if handled != tt.wantHandled || aborted != tt.wantHandled {
				t.Errorf("handled = %v, aborted = %v, want %v", handled, aborted, tt.wantHandled)
			}
			if invalidated := strings.Contains(header.Get("Set-Cookie"), pagetest.CookieName+"=;"); invalidated != tt.wantInvalidated {
				t.Errorf("cookie invalidated = %v, want %v (%q)", invalidated, tt.wantInvalidated, header.Get("Set-Cookie"))
			}
			if tt.wantHandled && (status != http.StatusUnauthorized || body == "") {
				t.Errorf("status = %d, body = %q", status, body)
			}
		})
	}
}

func TestHandleLatestTermsVersionMustBeAcceptedError(t *testing.T) {
	mustAccept := response.Error{Domain: termsErrors.Domain, Code: termsErrors.MustAcceptLatestTerms}
	tests := []struct {
		name        string
		statusCode  int
		apiError    response.Error
		wantHandled bool
	}{
		{"must accept", http.StatusPreconditionFailed, mustAccept, true},
		{"another status", http.StatusBadRequest, mustAccept, false},
		{"another domain", http.StatusPreconditionFailed, response.Error{Domain: "other", Code: termsErrors.MustAcceptLatestTerms}, false},
		{"another code", http.StatusPreconditionFailed, response.Error{Domain: termsErrors.Domain, Code: 999}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled, aborted, rec := handle(t, HandleLatestTermsVersionMustBeAcceptedError(), tt.statusCode, nil, tt.apiError)
			header, status := rec.Header(), rec.Code

			if handled != tt.wantHandled || aborted != tt.wantHandled {
				t.Errorf("handled = %v, aborted = %v, want %v", handled, aborted, tt.wantHandled)
			}
			if tt.wantHandled && (status != http.StatusFound || header.Get("Location") != pagetest.WebHost+"/en-us/account/terms") {
				t.Errorf("status = %d, location = %q", status, header.Get("Location"))
			}
		})
	}
}

func TestHandleAPIError(t *testing.T) {
	tests := []struct {
		name        string
		apiError    response.Error
		wantHandled bool
	}{
		{"error", response.Error{Code: 1, Domain: "test", Message: "It failed"}, true},
		{"no error", response.Error{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handled, aborted bool
			rec := serve(t, "en-us", nil, func(c *gin.Context) {
				handled = HandleAPIError(html.Main.PageByKey(c, "index"))(c, http.StatusBadRequest, nil, tt.apiError)
				aborted = c.IsAborted()
			})

			if handled != tt.wantHandled || aborted != tt.wantHandled {
				t.Errorf("handled = %v, aborted = %v, want %v", handled, aborted, tt.wantHandled)
			}
			if got := strings.Contains(rec.Body.String(), "It failed"); got != tt.wantHandled {
				t.Errorf("body = %q", rec.Body.String())
			}
		})
	}
}
