package api

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

func TestNewClient(t *testing.T) {
	var path, accept string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, accept = r.URL.Path, r.Header.Get("Accept")
	})
	setCookieConfig(t)
	c, _ := webtest.Context(http.MethodGet, "/", nil)

	client.Call(c, NewRequest())

	if path != "/v1" || accept != "application/json" {
		t.Errorf("path = %q, accept = %q", path, accept)
	}
}

func TestClient_Call_ForwardsTheRequest(t *testing.T) {
	setCookieConfig(t)

	tests := []struct {
		name          string
		authenticated bool
		cookie        string
		wantAuth      string
	}{
		{"authenticated", true, "token", "Bearer token"},
		{"authenticated without cookie", true, "", ""},
		{"not authenticated", false, "token", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var header http.Header
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { header = r.Header })

			c, _ := webtest.Context(http.MethodGet, "/", nil)
			context.SaveLocale("fr-fr")(c)
			c.Request.Header.Set("X-Request-ID", "request-1")
			c.Request.RemoteAddr = "1.2.3.4:1234"
			if tt.cookie != "" {
				c.Request.AddCookie(&http.Cookie{Name: "auth", Value: tt.cookie})
			}

			req := NewRequest()
			req.Authenticated = tt.authenticated
			client.Call(c, req)

			if got := header.Get("X-Api-Locale"); got != "fr-fr" {
				t.Errorf("locale = %q", got)
			}
			if got := header.Get("X-Request-ID"); got != "request-1" {
				t.Errorf("request ID = %q", got)
			}
			if got := header.Get("X-User-IP"); got != "1.2.3.4" {
				t.Errorf("user IP = %q", got)
			}
			if got := header.Get("Authorization"); got != tt.wantAuth {
				t.Errorf("authorization = %q, want %q", got, tt.wantAuth)
			}
		})
	}
}

func TestClient_Call_WithoutRequestID(t *testing.T) {
	var header http.Header
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { header = r.Header })
	c, _ := webtest.Context(http.MethodGet, "/", nil)

	client.Call(c, NewRequest())

	if _, ok := header["X-Request-Id"]; ok {
		t.Error("an absent request ID must not be forwarded")
	}
}

func TestClient_Call_BodyAndResponse(t *testing.T) {
	var method, contentType, body string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, contentType = r.Method, r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"done"}`))
	})
	c, _ := webtest.Context(http.MethodGet, "/", nil)

	var status Status
	req := NewRequest()
	req.Method = POST
	req.Endpoint = "/things"
	req.Body = map[string]string{"a": "b"}
	req.Response = &status
	req.ContentType = "application/json"

	result, ok := client.Call(c, req)

	if !ok || result.StatusCode != http.StatusOK || result.Handled || result.RequestError != nil || !result.APIError.IsNil() {
		t.Errorf("result = %+v, ok = %v", result, ok)
	}
	if method != http.MethodPost || contentType != "application/json" || body != `{"a":"b"}` {
		t.Errorf("method = %q, content type = %q, body = %q", method, contentType, body)
	}
	if status.Status != "done" {
		t.Errorf("response = %+v", status)
	}
}

func TestClient_Call_Methods(t *testing.T) {
	for _, method := range []string{GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS} {
		t.Run(method, func(t *testing.T) {
			var got string
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { got = r.Method })
			c, _ := webtest.Context(http.MethodGet, "/", nil)

			req := NewRequest()
			req.Method = method
			result, ok := client.Call(c, req)

			if got != method || !ok || result.StatusCode != http.StatusOK {
				t.Errorf("method = %q, status = %d, ok = %v", got, result.StatusCode, ok)
			}
		})
	}
}

func TestClient_Call_APIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(response.NewError(response.Error{Code: 12, Domain: "test", Message: "nope"}))
	})
	c, _ := webtest.Context(http.MethodGet, "/", nil)

	result, ok := client.Call(c, NewRequest())

	if ok || result.StatusCode != http.StatusBadRequest || result.APIError.Code != 12 || result.APIError.Message != "nope" || result.Handled {
		t.Errorf("result = %+v, ok = %v", result, ok)
	}
}

func TestClient_Call_ErrorStatusWithoutAPIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })
	c, _ := webtest.Context(http.MethodGet, "/", nil)

	result, ok := client.Call(c, NewRequest())

	if ok || result.StatusCode != http.StatusNotFound || !result.APIError.IsNil() {
		t.Errorf("result = %+v, ok = %v", result, ok)
	}
}

func TestClient_Call_RequestError(t *testing.T) {
	// Nothing listens on this port.
	client := NewClient("http://127.0.0.1", "v1", 1, 200, 0, false)
	c, _ := webtest.Context(http.MethodGet, "/", nil)

	var handlerError error
	req := NewRequest()
	req.ResponseHandlers = append(req.ResponseHandlers, func(c *gin.Context, statusCode int, requestError error, _ response.Error) bool {
		handlerError = requestError
		return false
	})

	result, ok := client.Call(c, req)

	if ok || result.StatusCode != 0 || result.RequestError == nil || handlerError == nil {
		t.Errorf("result = %+v, ok = %v, handler error = %v", result, ok, handlerError)
	}
}

func TestClient_Call_ResponseHandlers(t *testing.T) {
	setCookieConfig(t)

	var calls []string
	handler := func(name string, handles bool) ResponseHandlerFunc {
		return func(c *gin.Context, statusCode int, requestError error, apiError response.Error) bool {
			calls = append(calls, name)
			return handles
		}
	}

	tests := []struct {
		name        string
		useDefaults bool
		defaults    []ResponseHandlerFunc
		handlers    []ResponseHandlerFunc
		wantCalls   []string
		wantHandled bool
	}{
		{"defaults then request handlers", true, []ResponseHandlerFunc{handler("d1", false), handler("d2", false)}, []ResponseHandlerFunc{handler("r1", false)}, []string{"d1", "d2", "r1"}, false},
		{"a default handler stops the chain", true, []ResponseHandlerFunc{handler("d1", true), handler("d2", false)}, []ResponseHandlerFunc{handler("r1", false)}, []string{"d1"}, true},
		{"a request handler stops the chain", true, []ResponseHandlerFunc{handler("d1", false)}, []ResponseHandlerFunc{handler("r1", true), handler("r2", false)}, []string{"d1", "r1"}, true},
		{"defaults ignored", false, []ResponseHandlerFunc{handler("d1", true)}, []ResponseHandlerFunc{handler("r1", false)}, []string{"r1"}, false},
		{"no handler", true, nil, nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls = nil
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
			client.AddResponseHandlers(tt.defaults...)
			c, _ := webtest.Context(http.MethodGet, "/", nil)

			req := NewRequest()
			req.UseDefaultResponseHandlers = tt.useDefaults
			req.ResponseHandlers = tt.handlers
			result, ok := client.Call(c, req)

			if !equalStrings(calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", calls, tt.wantCalls)
			}
			if result.Handled != tt.wantHandled || ok == tt.wantHandled {
				t.Errorf("handled = %v, ok = %v", result.Handled, ok)
			}
		})
	}
}

func TestClient_Call_RenewedAuthToken(t *testing.T) {
	setCookieConfig(t)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(response.RenewedAuthTokenHeader, "renewed")
	})
	c, rec := webtest.Context(http.MethodGet, "/", nil)

	client.Call(c, NewRequest())

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "auth" || cookies[0].Value != "renewed" {
		t.Errorf("cookies = %v", cookies)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
