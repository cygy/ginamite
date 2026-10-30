// Package testing provides helpers to test the endpoints of a running API.
// The API address comes from the environment variables API_HOST and API_PORT.
package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	authErrors "github.com/cygy/ginamite/common/errors/auth"
	localeErrors "github.com/cygy/ginamite/common/errors/locale"
	requestErrors "github.com/cygy/ginamite/common/errors/request"
	r "github.com/cygy/ginamite/common/response"
)

// HTTP methods.
const (
	// GET is the HTTP method GET.
	GET = "GET"

	// POST is the HTTP method POST.
	POST = "POST"

	// PUT is the HTTP method PUT.
	PUT = "PUT"

	// PATCH is the HTTP method PATCH.
	PATCH = "PATCH"

	// DELETE is the HTTP method DELETE.
	DELETE = "DELETE"

	// HEAD is the HTTP method HEAD.
	HEAD = "HEAD"

	// OPTIONS is the HTTP method OPTIONS.
	OPTIONS = "OPTIONS"
)

var (
	apiAddress = os.Getenv("API_HOST") + ":" + os.Getenv("API_PORT")

	// APIVersion is the version of the API to test, used as the first segment of the endpoints.
	APIVersion = ""
)

// H is a shorthand for a map holding a JSON body.
type H map[string]any

// S is a shorthand for a map of strings, used for the HTTP headers.
type S map[string]string

// Reporter reports the failures of a test. It is implemented by *testing.T.
type Reporter interface {
	Error(args ...any)
	Errorf(format string, args ...any)
}

// Request is a request to send to the API, with the response it expects.
type Request struct {
	T                Reporter
	Method           string
	Endpoint         string
	Parameter        string // The query string of a GET request.
	Body             H      // The JSON body of a POST or PUT request.
	AuthToken        string
	RenewedAuthToken string // Set by Do with the renewed auth token received, if any.
	Headers          S
	ExpectedResponse struct {
		Code             int
		Status           string
		Error            r.Error
		RenewedAuthToken bool
	}
	Response any // Receives the decoded body of a successful response, if not nil.
}

// NewRequest returns a GET request expecting the status code 200.
func NewRequest(t *testing.T) Request {
	request := Request{
		T:      t,
		Method: GET,
	}

	request.ExpectedResponse.Code = http.StatusOK

	return request
}

// Do sends the request to the API and reports to T each difference with the
// expected response: status code, status message, error and renewed auth token.
// The response is decoded into Response when the status code is a success.
func (request *Request) Do() {
	var body []byte
	if (request.Method == POST || request.Method == PUT) && request.Body != nil {
		b, err := json.Marshal(request.Body)
		if err != nil {
			request.T.Error("Can not marshal the JSON body")
			return
		}
		body = b
	}

	url := fmt.Sprintf("http://%s", apiAddress)
	if len(APIVersion) > 0 {
		url = fmt.Sprintf("%s/%s", url, APIVersion)
	}
	url = fmt.Sprintf("%s%s", url, request.Endpoint)
	if request.Method == GET && len(request.Parameter) > 0 {
		url = fmt.Sprintf("%s?%s", url, request.Parameter)
	}

	httpRequest, err := http.NewRequest(request.Method, url, bytes.NewBuffer(body))
	if err != nil {
		request.T.Error(err)
		return
	}
	if len(body) > 0 {
		httpRequest.Header.Set("Content-Type", "application/json")
	}
	if len(request.AuthToken) > 0 {
		httpRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", request.AuthToken))
	}
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		request.T.Error(err)
		return
	}

	defer func() {
		if err := httpResponse.Body.Close(); err != nil {
			request.T.Error(err)
		}
	}()
	responseContent, _ := io.ReadAll(httpResponse.Body)

	logPrefix := fmt.Sprintf("[%s %s]", request.Method, request.Endpoint)

	if httpResponse.StatusCode != request.ExpectedResponse.Code {
		request.T.Errorf("%s code %d expected, but get %d (%s).", logPrefix, request.ExpectedResponse.Code, httpResponse.StatusCode, string(responseContent))
	}

	if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300 && request.Response != nil {
		if err := json.Unmarshal(responseContent, request.Response); err != nil {
			request.T.Errorf("%s unable to decode the response: %v", logPrefix, err)
		}
	}

	if len(request.ExpectedResponse.Status) > 0 || !request.ExpectedResponse.Error.IsNil() {
		var response r.Response
		if err := json.Unmarshal(responseContent, &response); err != nil {
			request.T.Errorf("%s unable to decode the response: %v", logPrefix, err)
		}

		if len(request.ExpectedResponse.Status) > 0 && request.ExpectedResponse.Status != response.Status {
			request.T.Errorf("%s status %s expected, but get %s.", logPrefix, request.ExpectedResponse.Status, response.Status)
		}

		if !request.ExpectedResponse.Error.IsNil() {
			if response.Error == nil {
				request.T.Errorf("%s error (%s: %d) expected, but get no error.", logPrefix, request.ExpectedResponse.Error.Domain, request.ExpectedResponse.Error.Code)
			} else if response.Error.Code != request.ExpectedResponse.Error.Code || response.Error.Domain != request.ExpectedResponse.Error.Domain {
				request.T.Errorf("%s error (%s: %d) expected, but get (%s: %d).", logPrefix, request.ExpectedResponse.Error.Domain, request.ExpectedResponse.Error.Code, response.Error.Domain, response.Error.Code)
			}
		}

		renewedAuthToken := httpResponse.Header.Get(r.RenewedAuthTokenHeader)
		if len(renewedAuthToken) > 0 {
			request.RenewedAuthToken = renewedAuthToken

			if !request.ExpectedResponse.RenewedAuthToken {
				request.T.Errorf("%s unexpected renewed auth token, but get one.", logPrefix)
			}
		}
		if len(renewedAuthToken) == 0 && request.ExpectedResponse.RenewedAuthToken {
			request.T.Errorf("%s expected renewed auth token, but get none.", logPrefix)
		}
	}
}

// OK verifies that the endpoint returns the HTTP status OK.
func (request *Request) OK() {
	request.ExpectedResponse.Code = http.StatusOK
	request.Do()
}

// Created verifies that the endpoint returns the HTTP status CREATED.
func (request *Request) Created() {
	request.ExpectedResponse.Code = http.StatusCreated
	request.Do()
}

// NoContent verifies that the endpoint returns the HTTP status NO CONTENT.
func (request *Request) NoContent() {
	request.ExpectedResponse.Code = http.StatusNoContent
	request.Do()
}

// Forbidden verifies that the endpoint returns the HTTP status FORBIDDEN.
func (request *Request) Forbidden() {
	request.ExpectedResponse.Code = http.StatusForbidden
	request.Do()
}

// UnauthorizedWithoutAuthorization removes the auth token and verifies that the endpoint returns the HTTP status UNAUTHORIZED.
func (request *Request) UnauthorizedWithoutAuthorization() {
	request.ExpectedResponse.Code = http.StatusUnauthorized
	request.AuthToken = ""
	request.Do()
}

// UnauthorizedWithoutRights verifies that the endpoint returns the HTTP status UNAUTHORIZED with the error "insufficient rights".
func (request *Request) UnauthorizedWithoutRights() {
	request.ExpectedResponse.Code = http.StatusUnauthorized
	request.ExpectedResponse.Error.Code = authErrors.InsufficientRights
	request.ExpectedResponse.Error.Domain = authErrors.Domain
	request.Do()
}

// Unauthorized verifies that the endpoint returns the HTTP status UNAUTHORIZED.
func (request *Request) Unauthorized() {
	request.ExpectedResponse.Code = http.StatusUnauthorized
	request.Do()
}

// NotFound verifies that the endpoint returns the HTTP status NOT FOUND.
func (request *Request) NotFound() {
	request.ExpectedResponse.Code = http.StatusNotFound
	request.Do()
}

// BadRequest verifies that the endpoint returns the HTTP status BAD REQUEST.
func (request *Request) BadRequest() {
	request.ExpectedResponse.Code = http.StatusBadRequest
	request.Do()
}

// PreconditionFailed verifies that the endpoint returns the HTTP status PRECONDITION FAILED.
func (request *Request) PreconditionFailed() {
	request.ExpectedResponse.Code = http.StatusPreconditionFailed
	request.Do()
}

// InvalidParameter verifies that the endpoint returns the HTTP status BAD REQUEST with the error "invalid parameter value".
func (request *Request) InvalidParameter() {
	request.ExpectedResponse.Error.Code = requestErrors.InvalidParameterValue
	request.ExpectedResponse.Error.Domain = requestErrors.Domain
	request.BadRequest()
}

// InvalidRequest verifies that the endpoint returns the HTTP status BAD REQUEST with the error "invalid request".
func (request *Request) InvalidRequest() {
	request.ExpectedResponse.Error.Code = requestErrors.InvalidRequest
	request.ExpectedResponse.Error.Domain = requestErrors.Domain
	request.BadRequest()
}

// MandatoryParameters sends a request for each body and verifies that the endpoint returns
// the HTTP status BAD REQUEST with the error "parameter not found". Each body lacks a mandatory parameter.
func (request *Request) MandatoryParameters(bodies []H) {
	request.ExpectedResponse.Error.Code = requestErrors.NotFoundParameterValue
	request.ExpectedResponse.Error.Domain = requestErrors.Domain

	for _, body := range bodies {
		request.Body = body
		request.BadRequest()
	}
}

// UnsupportedLocale verifies that the endpoint returns the HTTP status BAD REQUEST with the error "unsupported locale".
func (request *Request) UnsupportedLocale() {
	request.ExpectedResponse.Error.Code = localeErrors.UnsupportedLocale
	request.ExpectedResponse.Error.Domain = localeErrors.Domain
	request.BadRequest()
}
