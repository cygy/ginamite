package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authErrors "github.com/cygy/ginamite/common/errors/auth"
	localeErrors "github.com/cygy/ginamite/common/errors/locale"
	requestErrors "github.com/cygy/ginamite/common/errors/request"
	"github.com/cygy/ginamite/common/response"
)

// recorder is a Reporter keeping the failures it receives.
type recorder struct {
	errors []string
}

func (rec *recorder) Helper() {}

func (rec *recorder) Error(args ...any) { rec.errors = append(rec.errors, fmt.Sprint(args...)) }

func (rec *recorder) Errorf(format string, args ...any) {
	rec.errors = append(rec.errors, fmt.Sprintf(format, args...))
}

// seen is what the fake API received.
type seen struct {
	method, uri, auth, contentType, custom, body string
}

// serve starts a fake API answering with the status, the body and the header, and returns the API targeting it with
// the version "v1".
func serve(t *testing.T, status int, body string, header http.Header) (*API, *seen) {
	t.Helper()

	got := &seen{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		payload, _ := io.ReadAll(req.Body)
		*got = seen{
			method:      req.Method,
			uri:         req.URL.RequestURI(),
			auth:        req.Header.Get("Authorization"),
			contentType: req.Header.Get("Content-Type"),
			custom:      req.Header.Get("X-Custom"),
			body:        string(payload),
		}

		for k, v := range header {
			w.Header()[k] = v
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return &API{URL: server.URL + "/v1/", Client: server.Client()}, got
}

func errorBody(t *testing.T, domain string, code int) string {
	t.Helper()

	data, err := json.Marshal(response.Response{Error: &response.Error{Domain: domain, Code: code}})
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func TestFromEnv(t *testing.T) {
	t.Setenv("API_HOST", "api")
	t.Setenv("API_PORT", "80")

	if got := FromEnv("v1").URL; got != "http://api:80/v1" {
		t.Errorf("FromEnv(v1).URL = %q", got)
	}
	if got := FromEnv("").URL; got != "http://api:80" {
		t.Errorf("FromEnv().URL = %q", got)
	}
}

func TestAPI_NewRequest(t *testing.T) {
	api := &API{}
	request := api.NewRequest(t)

	if request.T != Reporter(t) || request.Method != http.MethodGet || request.Expected.Code != http.StatusOK || request.api != api {
		t.Errorf("NewRequest() = %+v", request)
	}
}

func TestRequest_Do(t *testing.T) {
	api, got := serve(t, http.StatusOK, `{"status":"done","id":"42"}`, nil)

	var decoded response.Response
	rec := &recorder{}
	request := api.NewRequest(rec)
	request.Endpoint = "/users"
	request.Query = "page=2"
	request.AuthToken = "tok"
	request.Headers = S{"X-Custom": "yes"}
	request.Response = &decoded
	request.Expected.Status = "done"
	request.Do()

	if len(rec.errors) != 0 {
		t.Fatalf("unexpected failures: %v", rec.errors)
	}
	if *got != (seen{method: http.MethodGet, uri: "/v1/users?page=2", auth: "Bearer tok", custom: "yes"}) {
		t.Errorf("request = %+v", got)
	}
	if decoded.ID != "42" {
		t.Errorf("decoded = %+v", decoded)
	}
}

func TestRequest_Do_Body(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			api, got := serve(t, http.StatusOK, `{}`, nil)

			rec := &recorder{}
			request := api.NewRequest(rec)
			request.Method = method
			request.Endpoint = "/users"
			request.Body = H{"name": "John"}
			request.Do()

			if len(rec.errors) != 0 {
				t.Fatalf("unexpected failures: %v", rec.errors)
			}
			if got.method != method || got.body != `{"name":"John"}` || got.contentType != "application/json" {
				t.Errorf("request = %+v", got)
			}
		})
	}
}

func TestRequest_Do_Failures(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		header http.Header
		setup  func(*Request)
		want   string
	}{
		{"unexpected code", http.StatusBadRequest, `bad`, nil, func(*Request) {}, "code 200 expected, got 400 (bad)"},
		{"undecodable response", http.StatusOK, `nope`, nil, func(q *Request) { q.Response = &response.Response{} }, "unable to decode the response"},
		{"undecodable response for the status", http.StatusOK, `nope`, nil, func(q *Request) { q.Expected.Status = "done" }, "unable to decode the response"},
		{"wrong status", http.StatusOK, `{"status":"other"}`, nil, func(q *Request) { q.Expected.Status = "done" }, "status done expected, got other"},
		{
			"missing error", http.StatusOK, `{}`, nil,
			func(q *Request) { q.Expected.Error = response.Error{Code: 1, Domain: "d"} },
			"error (d: 1) expected, got no error",
		},
		{
			"wrong error", http.StatusOK, `{"error":{"code":2,"domain":"d"}}`, nil,
			func(q *Request) { q.Expected.Error = response.Error{Code: 1, Domain: "d"} },
			"error (d: 1) expected, got (d: 2)",
		},
		{
			"unexpected renewed token", http.StatusOK, `{}`,
			http.Header{response.RenewedAuthTokenHeader: {"new"}},
			func(*Request) {}, "renewed auth token expected: false, got: true",
		},
		{
			"missing renewed token", http.StatusOK, `{}`, nil,
			func(q *Request) { q.Expected.RenewedAuthToken = true }, "renewed auth token expected: true, got: false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api, _ := serve(t, tt.status, tt.body, tt.header)

			rec := &recorder{}
			request := api.NewRequest(rec)
			tt.setup(request)
			request.Do()

			if !strings.Contains(strings.Join(rec.errors, "\n"), tt.want) {
				t.Errorf("failures = %v, want one containing %q", rec.errors, tt.want)
			}
		})
	}
}

func TestRequest_Do_RenewedAuthToken(t *testing.T) {
	api, _ := serve(t, http.StatusOK, `{}`, http.Header{response.RenewedAuthTokenHeader: {"new"}})

	rec := &recorder{}
	request := api.NewRequest(rec)
	request.Expected.RenewedAuthToken = true
	request.Do()

	if len(rec.errors) != 0 || request.RenewedAuthToken != "new" {
		t.Errorf("failures = %v, token = %q", rec.errors, request.RenewedAuthToken)
	}
}

func TestRequest_Do_TransportFailures(t *testing.T) {
	t.Run("unencodable body", func(t *testing.T) {
		rec := &recorder{}
		request := (&API{}).NewRequest(rec)
		request.Body = H{"bad": make(chan int)}
		request.Do()

		if len(rec.errors) != 1 || !strings.Contains(rec.errors[0], "encode") {
			t.Errorf("failures = %v", rec.errors)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		rec := &recorder{}
		request := (&API{}).NewRequest(rec)
		request.Method = "bad method"
		request.Do()

		if len(rec.errors) != 1 {
			t.Errorf("failures = %v", rec.errors)
		}
	})

	t.Run("unreachable API", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		server.Close()

		rec := &recorder{}
		(&API{URL: server.URL}).NewRequest(rec).Do()

		if len(rec.errors) != 1 {
			t.Errorf("failures = %v", rec.errors)
		}
	})
}

func TestRequest_Shortcuts(t *testing.T) {
	tests := []struct {
		name   string
		call   func(*Request)
		status int
		domain string
		code   int
	}{
		{"OK", (*Request).OK, http.StatusOK, "", 0},
		{"Created", (*Request).Created, http.StatusCreated, "", 0},
		{"NoContent", (*Request).NoContent, http.StatusNoContent, "", 0},
		{"Forbidden", (*Request).Forbidden, http.StatusForbidden, "", 0},
		{"Unauthorized", (*Request).Unauthorized, http.StatusUnauthorized, "", 0},
		{"UnauthorizedWithoutAuthorization", (*Request).UnauthorizedWithoutAuthorization, http.StatusUnauthorized, "", 0},
		{"NotFound", (*Request).NotFound, http.StatusNotFound, "", 0},
		{"BadRequest", (*Request).BadRequest, http.StatusBadRequest, "", 0},
		{"PreconditionFailed", (*Request).PreconditionFailed, http.StatusPreconditionFailed, "", 0},
		{"UnauthorizedWithoutRights", (*Request).UnauthorizedWithoutRights, http.StatusUnauthorized, authErrors.Domain, authErrors.InsufficientRights},
		{"InvalidParameter", (*Request).InvalidParameter, http.StatusBadRequest, requestErrors.Domain, requestErrors.InvalidParameterValue},
		{"InvalidRequest", (*Request).InvalidRequest, http.StatusBadRequest, requestErrors.Domain, requestErrors.InvalidRequest},
		{"UnsupportedLocale", (*Request).UnsupportedLocale, http.StatusBadRequest, localeErrors.Domain, localeErrors.UnsupportedLocale},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := "{}"
			if tt.domain != "" {
				body = errorBody(t, tt.domain, tt.code)
			}
			api, got := serve(t, tt.status, body, nil)

			rec := &recorder{}
			request := api.NewRequest(rec)
			request.AuthToken = "tok"
			tt.call(request)

			if len(rec.errors) != 0 {
				t.Errorf("unexpected failures: %v", rec.errors)
			}
			if wantAuth := tt.name != "UnauthorizedWithoutAuthorization"; (got.auth != "") != wantAuth {
				t.Errorf("authorization = %q", got.auth)
			}
		})
	}
}

func TestRequest_Shortcuts_ReportMismatch(t *testing.T) {
	api, _ := serve(t, http.StatusOK, `{}`, nil)

	rec := &recorder{}
	api.NewRequest(rec).NotFound()

	if len(rec.errors) != 1 {
		t.Errorf("failures = %v", rec.errors)
	}
}

func TestRequest_MandatoryParameters(t *testing.T) {
	api, got := serve(t, http.StatusBadRequest, errorBody(t, requestErrors.Domain, requestErrors.NotFoundParameterValue), nil)

	rec := &recorder{}
	request := api.NewRequest(rec)
	request.Method = http.MethodPost
	request.MandatoryParameters([]H{{"a": 1}, {"b": 2}})

	if len(rec.errors) != 0 {
		t.Errorf("unexpected failures: %v", rec.errors)
	}
	if got.body != `{"b":2}` {
		t.Errorf("last body = %q", got.body)
	}
}
