package testing

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
	r "github.com/cygy/ginamite/common/response"
)

// recorder is a Reporter keeping the failures it receives.
type recorder struct {
	errors []string
}

func (rec *recorder) Error(args ...any) { rec.errors = append(rec.errors, fmt.Sprint(args...)) }

func (rec *recorder) Errorf(format string, args ...any) {
	rec.errors = append(rec.errors, fmt.Sprintf(format, args...))
}

// seen is what the fake API received.
type seen struct {
	method, uri, auth, contentType, custom, body string
}

// serve starts a fake API answering with the status and the body, and targets it.
func serve(t *testing.T, status int, body string, header http.Header) *seen {
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

	oldAddress, oldVersion := apiAddress, APIVersion
	apiAddress = server.Listener.Addr().String()
	t.Cleanup(func() {
		server.Close()
		apiAddress, APIVersion = oldAddress, oldVersion
	})

	return got
}

func newRequest(rec *recorder) *Request {
	request := NewRequest(&testing.T{})
	request.T = rec

	return &request
}

func TestNewRequest(t *testing.T) {
	request := NewRequest(t)

	if request.T != Reporter(t) || request.Method != GET || request.ExpectedResponse.Code != http.StatusOK {
		t.Errorf("NewRequest() = %+v", request)
	}
}

func TestRequest_Do_Success(t *testing.T) {
	got := serve(t, http.StatusOK, `{"status":"done","id":"42"}`, nil)
	APIVersion = "v1"

	var decoded r.Response
	rec := &recorder{}
	request := newRequest(rec)
	request.Endpoint = "/users"
	request.Parameter = "page=2"
	request.AuthToken = "tok"
	request.Headers = S{"X-Custom": "yes"}
	request.Response = &decoded
	request.ExpectedResponse.Status = "done"
	request.Do()

	if len(rec.errors) != 0 {
		t.Fatalf("unexpected failures: %v", rec.errors)
	}
	if got.method != GET || got.uri != "/v1/users?page=2" || got.auth != "Bearer tok" || got.custom != "yes" || got.body != "" {
		t.Errorf("request = %+v", got)
	}
	if decoded.ID != "42" {
		t.Errorf("decoded = %+v", decoded)
	}
}

func TestRequest_Do_Body(t *testing.T) {
	for _, method := range []string{POST, PUT, PATCH} {
		t.Run(method, func(t *testing.T) {
			got := serve(t, http.StatusOK, `{}`, nil)

			rec := &recorder{}
			request := newRequest(rec)
			request.Method = method
			request.Endpoint = "/users"
			request.Parameter = "ignored=1"
			request.Body = H{"name": "John"}
			request.Do()

			if len(rec.errors) != 0 {
				t.Fatalf("unexpected failures: %v", rec.errors)
			}

			wantBody := `{"name":"John"}`
			if method == PATCH {
				wantBody = ""
			}
			if got.method != method || got.uri != "/users" || got.body != wantBody {
				t.Errorf("request = %+v", got)
			}
			if method != PATCH && got.contentType != "application/json" {
				t.Errorf("content type = %q", got.contentType)
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
		{"unexpected code", http.StatusBadRequest, `bad`, nil, func(*Request) {}, "code 200 expected, but get 400 (bad)"},
		{"undecodable response", http.StatusOK, `nope`, nil, func(q *Request) { q.Response = &r.Response{} }, "unable to decode the response"},
		{"undecodable response for the status", http.StatusOK, `nope`, nil, func(q *Request) { q.ExpectedResponse.Status = "done" }, "unable to decode the response"},
		{"wrong status", http.StatusOK, `{"status":"other"}`, nil, func(q *Request) { q.ExpectedResponse.Status = "done" }, "status done expected, but get other"},
		{
			"missing error", http.StatusOK, `{}`, nil,
			func(q *Request) { q.ExpectedResponse.Error = r.Error{Code: 1, Domain: "d"} },
			"error (d: 1) expected, but get no error",
		},
		{
			"wrong error", http.StatusOK, `{"error":{"code":2,"domain":"d"}}`, nil,
			func(q *Request) { q.ExpectedResponse.Error = r.Error{Code: 1, Domain: "d"} },
			"error (d: 1) expected, but get (d: 2)",
		},
		{
			"unexpected renewed token", http.StatusOK, `{"status":"done"}`,
			http.Header{r.RenewedAuthTokenHeader: {"new"}},
			func(q *Request) { q.ExpectedResponse.Status = "done" },
			"unexpected renewed auth token",
		},
		{
			"missing renewed token", http.StatusOK, `{"status":"done"}`, nil,
			func(q *Request) { q.ExpectedResponse.Status = "done"; q.ExpectedResponse.RenewedAuthToken = true },
			"expected renewed auth token, but get none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serve(t, tt.status, tt.body, tt.header)

			rec := &recorder{}
			request := newRequest(rec)
			tt.setup(request)
			request.Do()

			if !strings.Contains(strings.Join(rec.errors, "\n"), tt.want) {
				t.Errorf("failures = %v, want one containing %q", rec.errors, tt.want)
			}
		})
	}
}

func TestRequest_Do_RenewedAuthToken(t *testing.T) {
	serve(t, http.StatusOK, `{"status":"done"}`, http.Header{r.RenewedAuthTokenHeader: {"new"}})

	rec := &recorder{}
	request := newRequest(rec)
	request.ExpectedResponse.Status = "done"
	request.ExpectedResponse.RenewedAuthToken = true
	request.Do()

	if len(rec.errors) != 0 || request.RenewedAuthToken != "new" {
		t.Errorf("failures = %v, token = %q", rec.errors, request.RenewedAuthToken)
	}
}

func TestRequest_Do_TransportFailures(t *testing.T) {
	t.Run("unmarshalable body", func(t *testing.T) {
		rec := &recorder{}
		request := newRequest(rec)
		request.Method = POST
		request.Body = H{"bad": make(chan int)}
		request.Do()

		if len(rec.errors) != 1 || !strings.Contains(rec.errors[0], "marshal") {
			t.Errorf("failures = %v", rec.errors)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		rec := &recorder{}
		request := newRequest(rec)
		request.Method = "bad method"
		request.Do()

		if len(rec.errors) != 1 {
			t.Errorf("failures = %v", rec.errors)
		}
	})

	t.Run("unreachable API", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		address := server.Listener.Addr().String()
		server.Close()

		old := apiAddress
		apiAddress = address
		defer func() { apiAddress = old }()

		rec := &recorder{}
		newRequest(rec).Do()

		if len(rec.errors) != 1 {
			t.Errorf("failures = %v", rec.errors)
		}
	})
}

func TestRequest_Shortcuts(t *testing.T) {
	type shortcut struct {
		name      string
		call      func(*Request)
		status    int
		errorBody string
		check     func(*testing.T, *Request)
	}

	errorBody := func(domain string, code int) string {
		data, _ := json.Marshal(r.Response{Error: &r.Error{Domain: domain, Code: code}})
		return string(data)
	}

	tests := []shortcut{
		{"OK", (*Request).OK, http.StatusOK, "", nil},
		{"Created", (*Request).Created, http.StatusCreated, "", nil},
		{"NoContent", (*Request).NoContent, http.StatusNoContent, "", nil},
		{"Forbidden", (*Request).Forbidden, http.StatusForbidden, "", nil},
		{"Unauthorized", (*Request).Unauthorized, http.StatusUnauthorized, "", nil},
		{"NotFound", (*Request).NotFound, http.StatusNotFound, "", nil},
		{"BadRequest", (*Request).BadRequest, http.StatusBadRequest, "", nil},
		{"PreconditionFailed", (*Request).PreconditionFailed, http.StatusPreconditionFailed, "", nil},
		{
			"UnauthorizedWithoutAuthorization", (*Request).UnauthorizedWithoutAuthorization, http.StatusUnauthorized, "",
			func(t *testing.T, q *Request) {
				t.Helper()
				if q.AuthToken != "" {
					t.Error("the auth token must be removed")
				}
			},
		},
		{"UnauthorizedWithoutRights", (*Request).UnauthorizedWithoutRights, http.StatusUnauthorized, errorBody(authErrors.Domain, authErrors.InsufficientRights), nil},
		{"InvalidParameter", (*Request).InvalidParameter, http.StatusBadRequest, errorBody(requestErrors.Domain, requestErrors.InvalidParameterValue), nil},
		{"InvalidRequest", (*Request).InvalidRequest, http.StatusBadRequest, errorBody(requestErrors.Domain, requestErrors.InvalidRequest), nil},
		{"UnsupportedLocale", (*Request).UnsupportedLocale, http.StatusBadRequest, errorBody(localeErrors.Domain, localeErrors.UnsupportedLocale), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.errorBody
			if body == "" {
				body = "{}"
			}
			serve(t, tt.status, body, nil)

			rec := &recorder{}
			request := newRequest(rec)
			request.AuthToken = "tok"
			tt.call(request)

			if len(rec.errors) != 0 {
				t.Errorf("unexpected failures: %v", rec.errors)
			}
			if tt.check != nil {
				tt.check(t, request)
			}
		})
	}
}

func TestRequest_Shortcuts_ReportMismatch(t *testing.T) {
	serve(t, http.StatusOK, `{}`, nil)

	rec := &recorder{}
	newRequest(rec).NotFound()

	if len(rec.errors) != 1 {
		t.Errorf("failures = %v", rec.errors)
	}
}

func TestRequest_MandatoryParameters(t *testing.T) {
	got := serve(t, http.StatusBadRequest, string(mustJSON(t, r.Response{Error: &r.Error{Domain: requestErrors.Domain, Code: requestErrors.NotFoundParameterValue}})), nil)

	rec := &recorder{}
	request := newRequest(rec)
	request.Method = POST
	request.MandatoryParameters([]H{{"a": 1}, {"b": 2}})

	if len(rec.errors) != 0 {
		t.Errorf("unexpected failures: %v", rec.errors)
	}
	if got.body != `{"b":2}` {
		t.Errorf("last body = %q", got.body)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return data
}
