package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	authErrors "github.com/cygy/ginamite/common/errors/auth"
	localeErrors "github.com/cygy/ginamite/common/errors/locale"
	requestErrors "github.com/cygy/ginamite/common/errors/request"
	r "github.com/cygy/ginamite/common/response"
)

// Do : executes the request
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

	httpRequest, _ := http.NewRequest(request.Method, url, bytes.NewBuffer(body))
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

	defer httpResponse.Body.Close()
	responseContent, _ := io.ReadAll(httpResponse.Body)

	logPrefix := fmt.Sprintf("[%s %s]", request.Method, request.Endpoint)

	if httpResponse.StatusCode != request.ExpectedResponse.Code {
		request.T.Errorf("%s code %d expected, but get %d (%s).", logPrefix, request.ExpectedResponse.Code, httpResponse.StatusCode, string(responseContent))
	}

	if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300 && request.Response != nil {
		json.Unmarshal(responseContent, request.Response)
	}

	if len(request.ExpectedResponse.Status) > 0 || !request.ExpectedResponse.Error.IsNil() {
		var response r.Response
		json.Unmarshal(responseContent, &response)

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

// OK : verifies that the endpoint returns a http status OK.
func (request *Request) OK() {
	request.ExpectedResponse.Code = http.StatusOK
	request.Do()
}

// Created : verifies that the endpoint returns a http status CREATED.
func (request *Request) Created() {
	request.ExpectedResponse.Code = http.StatusCreated
	request.Do()
}

// NoContent : verifies that the endpoint returns a http status NO CONTENT.
func (request *Request) NoContent() {
	request.ExpectedResponse.Code = http.StatusNoContent
	request.Do()
}

// Forbidden : verifies that the endpoint returns a http status FORBIDDEN.
func (request *Request) Forbidden() {
	request.ExpectedResponse.Code = http.StatusForbidden
	request.Do()
}

// UnauthorizedWithoutAuthorization : verifies that the endpoint returns a http status UNAUTHORIZED.
func (request *Request) UnauthorizedWithoutAuthorization() {
	request.ExpectedResponse.Code = http.StatusUnauthorized
	request.AuthToken = ""
	request.Do()
}

// UnauthorizedWithoutRights : verifies that the endpoint returns a http status UNAUTHORIZED.
func (request *Request) UnauthorizedWithoutRights() {
	request.ExpectedResponse.Code = http.StatusUnauthorized
	request.ExpectedResponse.Error.Code = authErrors.InsufficientRights
	request.ExpectedResponse.Error.Domain = authErrors.Domain
	request.Do()
}

// Unauthorized : verifies that the endpoint returns a http status UNAUTHORIZED.
func (request *Request) Unauthorized() {
	request.ExpectedResponse.Code = http.StatusUnauthorized
	request.Do()
}

// NotFound : verifies that the endpoint returns a http status NOT FOUND.
func (request *Request) NotFound() {
	request.ExpectedResponse.Code = http.StatusNotFound
	request.Do()
}

// BadRequest : verifies that the endpoint returns a http status BAD REQUEST.
func (request *Request) BadRequest() {
	request.ExpectedResponse.Code = http.StatusBadRequest
	request.Do()
}

// PreconditionFailed : verifies that the endpoint returns a http status PRECONDITION FAILED.
func (request *Request) PreconditionFailed() {
	request.ExpectedResponse.Code = http.StatusPreconditionFailed
	request.Do()
}

// InvalidParameter : verifies that the endpoint returns a http status BAD REQUEST.
func (request *Request) InvalidParameter() {
	request.ExpectedResponse.Error.Code = requestErrors.InvalidParameterValue
	request.ExpectedResponse.Error.Domain = requestErrors.Domain
	request.BadRequest()
}

// InvalidRequest : verifies that the endpoint returns a http status BAD REQUEST.
func (request *Request) InvalidRequest() {
	request.ExpectedResponse.Error.Code = requestErrors.InvalidRequest
	request.ExpectedResponse.Error.Domain = requestErrors.Domain
	request.BadRequest()
}

// MandatoryParameters : verifies that the endpoint returns a http status BAD REQUEST.
func (request *Request) MandatoryParameters(bodies []H) {
	request.ExpectedResponse.Error.Code = requestErrors.NotFoundParameterValue
	request.ExpectedResponse.Error.Domain = requestErrors.Domain

	for _, body := range bodies {
		request.Body = body
		request.BadRequest()
	}
}

// UnsupportedLocale : verifies that the endpoint returns a http status BAD REQUEST.
func (request *Request) UnsupportedLocale() {
	request.ExpectedResponse.Error.Code = localeErrors.UnsupportedLocale
	request.ExpectedResponse.Error.Domain = localeErrors.Domain
	request.BadRequest()
}
