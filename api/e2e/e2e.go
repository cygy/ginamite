// Package e2e tests the endpoints of a running API, end to end: each request is sent over HTTP, and its response is
// compared with the expected one.
//
//	var api = e2e.FromEnv("v1")
//
//	func TestHello(t *testing.T) {
//		req := api.NewRequest(t)
//		req.Endpoint = "/hello"
//		req.OK()
//	}
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	authErrors "github.com/cygy/ginamite/common/errors/auth"
	localeErrors "github.com/cygy/ginamite/common/errors/locale"
	requestErrors "github.com/cygy/ginamite/common/errors/request"
	"github.com/cygy/ginamite/common/response"
)

// H is a shorthand for a map holding a JSON body.
type H map[string]any

// S is a shorthand for a map of strings, used for the HTTP headers.
type S map[string]string

// Reporter reports the failures of a test. It is implemented by *testing.T.
type Reporter interface {
	Helper()
	Error(args ...any)
	Errorf(format string, args ...any)
}

// API is the API under test.
type API struct {
	// URL is the base URL of the endpoints, with the version of the API, for instance "http://localhost:8080/v1".
	URL string
	// Client sends the requests, http.DefaultClient if nil.
	Client *http.Client
}

// FromEnv returns the API at http://$API_HOST:$API_PORT/version, or without version if it is empty.
func FromEnv(version string) *API {
	url := "http://" + os.Getenv("API_HOST") + ":" + os.Getenv("API_PORT")
	if version != "" {
		url += "/" + version
	}

	return &API{URL: url}
}

// NewRequest returns a GET request to the API, expecting the status code 200.
func (api *API) NewRequest(t Reporter) *Request {
	return &Request{
		T:        t,
		Method:   http.MethodGet,
		Expected: Expected{Code: http.StatusOK},
		api:      api,
	}
}

// Expected is the response a request expects.
type Expected struct {
	Code             int            // The HTTP status code.
	Status           string         // The status of the body, not checked if empty.
	Error            response.Error // The error of the body, not checked if nil.
	RenewedAuthToken bool           // Whether the response renews the auth token.
}

// Request is a request to send to the API, with the response it expects. Create it with API.NewRequest.
type Request struct {
	T         Reporter
	Method    string // http.MethodGet by default.
	Endpoint  string // The path following the URL of the API, for instance "/users".
	Query     string // The query string, without the leading "?".
	Body      H      // Sent as JSON if not nil.
	AuthToken string // Sent as a bearer token if not empty.
	Headers   S
	Expected  Expected
	Response  any // Receives the decoded body of a successful response, if not nil.

	// RenewedAuthToken is set by Do with the renewed auth token received, if any.
	RenewedAuthToken string

	api *API
}

// Do sends the request to the API and reports to T each difference with the expected response: status code, status,
// error and renewed auth token. The body of a successful response is decoded into Response.
func (request *Request) Do() {
	request.T.Helper()

	body := io.Reader(http.NoBody)
	if request.Body != nil {
		data, err := json.Marshal(request.Body)
		if err != nil {
			request.T.Errorf("unable to encode the body: %v", err)
			return
		}
		body = bytes.NewReader(data)
	}

	url := strings.TrimSuffix(request.api.URL, "/") + request.Endpoint
	if request.Query != "" {
		url += "?" + request.Query
	}

	httpRequest, err := http.NewRequest(request.Method, url, body)
	if err != nil {
		request.T.Error(err)
		return
	}
	if request.Body != nil {
		httpRequest.Header.Set("Content-Type", "application/json")
	}
	if request.AuthToken != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+request.AuthToken)
	}
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}

	client := request.api.Client
	if client == nil {
		client = http.DefaultClient
	}

	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		request.T.Error(err)
		return
	}
	defer func() {
		if err := httpResponse.Body.Close(); err != nil {
			request.T.Error(err)
		}
	}()

	content, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		request.T.Errorf("[%s %s] unable to read the response: %v", request.Method, request.Endpoint, err)
		return
	}

	request.check(httpResponse, content)
}

// check reports each difference between the response and the expected one.
func (request *Request) check(httpResponse *http.Response, content []byte) {
	request.T.Helper()

	prefix := fmt.Sprintf("[%s %s]", request.Method, request.Endpoint)
	expected := request.Expected

	if httpResponse.StatusCode != expected.Code {
		request.T.Errorf("%s code %d expected, got %d (%s)", prefix, expected.Code, httpResponse.StatusCode, content)
	}

	if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300 && request.Response != nil {
		if err := json.Unmarshal(content, request.Response); err != nil {
			request.T.Errorf("%s unable to decode the response: %v", prefix, err)
		}
	}

	request.RenewedAuthToken = httpResponse.Header.Get(response.RenewedAuthTokenHeader)
	if renewed := request.RenewedAuthToken != ""; renewed != expected.RenewedAuthToken {
		request.T.Errorf("%s renewed auth token expected: %v, got: %v", prefix, expected.RenewedAuthToken, renewed)
	}

	if expected.Status == "" && expected.Error.IsNil() {
		return
	}

	var body response.Response
	if err := json.Unmarshal(content, &body); err != nil {
		request.T.Errorf("%s unable to decode the response: %v", prefix, err)
		return
	}

	if expected.Status != "" && expected.Status != body.Status {
		request.T.Errorf("%s status %s expected, got %s", prefix, expected.Status, body.Status)
	}

	if !expected.Error.IsNil() {
		switch {
		case body.Error == nil:
			request.T.Errorf("%s error (%s: %d) expected, got no error", prefix, expected.Error.Domain, expected.Error.Code)
		case body.Error.Code != expected.Error.Code || body.Error.Domain != expected.Error.Domain:
			request.T.Errorf("%s error (%s: %d) expected, got (%s: %d)", prefix, expected.Error.Domain, expected.Error.Code, body.Error.Domain, body.Error.Code)
		}
	}
}

// expect sets the expected status code and sends the request.
func (request *Request) expect(code int) {
	request.T.Helper()

	request.Expected.Code = code
	request.Do()
}

// expectError sets the expected status code and error, and sends the request.
func (request *Request) expectError(code int, domain string, errorCode int) {
	request.T.Helper()

	request.Expected.Error = response.Error{Domain: domain, Code: errorCode}
	request.expect(code)
}

// OK verifies that the endpoint returns the HTTP status OK.
func (request *Request) OK() {
	request.T.Helper()
	request.expect(http.StatusOK)
}

// Created verifies that the endpoint returns the HTTP status CREATED.
func (request *Request) Created() {
	request.T.Helper()
	request.expect(http.StatusCreated)
}

// NoContent verifies that the endpoint returns the HTTP status NO CONTENT.
func (request *Request) NoContent() {
	request.T.Helper()
	request.expect(http.StatusNoContent)
}

// Forbidden verifies that the endpoint returns the HTTP status FORBIDDEN.
func (request *Request) Forbidden() {
	request.T.Helper()
	request.expect(http.StatusForbidden)
}

// Unauthorized verifies that the endpoint returns the HTTP status UNAUTHORIZED.
func (request *Request) Unauthorized() {
	request.T.Helper()
	request.expect(http.StatusUnauthorized)
}

// UnauthorizedWithoutAuthorization removes the auth token and verifies that the endpoint returns the HTTP status
// UNAUTHORIZED.
func (request *Request) UnauthorizedWithoutAuthorization() {
	request.T.Helper()

	request.AuthToken = ""
	request.expect(http.StatusUnauthorized)
}

// UnauthorizedWithoutRights verifies that the endpoint returns the HTTP status UNAUTHORIZED with the error
// "insufficient rights".
func (request *Request) UnauthorizedWithoutRights() {
	request.T.Helper()
	request.expectError(http.StatusUnauthorized, authErrors.Domain, authErrors.InsufficientRights)
}

// NotFound verifies that the endpoint returns the HTTP status NOT FOUND.
func (request *Request) NotFound() {
	request.T.Helper()
	request.expect(http.StatusNotFound)
}

// BadRequest verifies that the endpoint returns the HTTP status BAD REQUEST.
func (request *Request) BadRequest() {
	request.T.Helper()
	request.expect(http.StatusBadRequest)
}

// PreconditionFailed verifies that the endpoint returns the HTTP status PRECONDITION FAILED.
func (request *Request) PreconditionFailed() {
	request.T.Helper()
	request.expect(http.StatusPreconditionFailed)
}

// InvalidParameter verifies that the endpoint returns the HTTP status BAD REQUEST with the error "invalid parameter
// value".
func (request *Request) InvalidParameter() {
	request.T.Helper()
	request.expectError(http.StatusBadRequest, requestErrors.Domain, requestErrors.InvalidParameterValue)
}

// InvalidRequest verifies that the endpoint returns the HTTP status BAD REQUEST with the error "invalid request".
func (request *Request) InvalidRequest() {
	request.T.Helper()
	request.expectError(http.StatusBadRequest, requestErrors.Domain, requestErrors.InvalidRequest)
}

// UnsupportedLocale verifies that the endpoint returns the HTTP status BAD REQUEST with the error "unsupported locale".
func (request *Request) UnsupportedLocale() {
	request.T.Helper()
	request.expectError(http.StatusBadRequest, localeErrors.Domain, localeErrors.UnsupportedLocale)
}

// MandatoryParameters sends the request once per body, each one lacking a mandatory parameter, and verifies that the
// endpoint returns the HTTP status BAD REQUEST with the error "parameter not found".
func (request *Request) MandatoryParameters(bodies []H) {
	request.T.Helper()

	for _, body := range bodies {
		request.Body = body
		request.expectError(http.StatusBadRequest, requestErrors.Domain, requestErrors.NotFoundParameterValue)
	}
}
