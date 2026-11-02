package api

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/log"
	r "github.com/cygy/ginamite/common/request"
	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/cookie"

	"github.com/gin-gonic/gin"
	resty "github.com/go-resty/resty/v2"
)

// ResponseHandlerFunc handles the response of the API to the request c, and reports whether it did:
// the next handlers are not called then.
// The status code is zero, and requestError is set, if the API did not respond.
type ResponseHandlerFunc func(c *gin.Context, statusCode int, requestError error, APIError response.Error) bool

// Client is a client of the API.
type Client struct {
	client           *resty.Client
	responseHandlers []ResponseHandlerFunc
}

// NewClient returns a client of the API at host:port/APIVersion, whose timeout is in milliseconds.
func NewClient(host, APIVersion string, port, timeout, retryCount int, debug bool) (client *Client) {
	hostURL := fmt.Sprintf("%s:%d/%s", host, port, APIVersion)

	apiClient := resty.New()
	apiClient.SetBaseURL(hostURL).
		SetHeader("Accept", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetTimeout(time.Duration(timeout) * time.Millisecond).
		SetRetryCount(retryCount).
		SetDebug(debug)

	client = &Client{
		client:           apiClient,
		responseHandlers: []ResponseHandlerFunc{},
	}
	return
}

// AddResponseHandlers adds some response handlers to the default ones.
func (client *Client) AddResponseHandlers(handlers ...ResponseHandlerFunc) {
	client.responseHandlers = append(client.responseHandlers, handlers...)
}

// Call sends the request to the API on behalf of the incoming request c: the locale, the request ID and the IP
// address of the user are forwarded, as well as the authentication cookie if the request is authenticated.
//
// The response is given to the default response handlers, if the request uses them, then to the handlers of the
// request, until one handles it. A renewed auth token returned by the API is saved to the authentication cookie.
// The boolean reports whether the call succeeded: the response was not handled by a handler, has no error, and has
// a status code under 300.
func (client *Client) Call(c *gin.Context, req Request) (Result, bool) {
	request := client.newRequest(c, req)

	var apiError struct {
		Error response.Error
	}
	request.SetError(&apiError)

	res, requestError := request.Execute(req.Method, req.Endpoint)

	statusCode := 0
	if res != nil {
		statusCode = res.StatusCode()
	}

	handlers := req.ResponseHandlers
	if req.UseDefaultResponseHandlers {
		handlers = append(append([]ResponseHandlerFunc{}, client.responseHandlers...), handlers...)
	}

	handled := false
	for _, handler := range handlers {
		if handler(c, statusCode, requestError, apiError.Error) {
			handled = true
			break
		}
	}

	// Any API call can return a renewed auth token by the headers.
	// It must be saved in a cookie.
	if res != nil {
		if renewedAuthToken := res.Header().Get(response.RenewedAuthTokenHeader); len(renewedAuthToken) > 0 {
			cookie.Create(c, renewedAuthToken)
		}
	}

	return Result{
		StatusCode:   statusCode,
		RequestError: requestError,
		APIError:     apiError.Error,
		Handled:      handled,
	}, !handled && requestError == nil && apiError.Error.IsNil() && statusCode < 300
}

// newRequest returns the request to send to the API, with the headers forwarded from the incoming request c.
func (client *Client) newRequest(c *gin.Context, req Request) *resty.Request {
	request := client.client.R().
		SetHeader(r.LocaleHeaderName, context.GetLocale(c)).
		SetHeader(r.UserIPAddressHeaderName, c.ClientIP())

	if requestID := c.Request.Header.Get(log.RequestIDHeaderName); len(requestID) > 0 {
		request.SetHeader(log.RequestIDHeaderName, requestID)
	}

	if req.Authenticated {
		if token, _ := cookie.GetValue(c); len(token) > 0 {
			request.SetAuthToken(token)
		}
	}

	if len(req.ContentType) > 0 {
		request.SetHeader("Content-Type", req.ContentType)
	}

	if req.Body != nil {
		request.SetBody(req.Body)
	}

	if req.Response != nil {
		request.SetResult(&req.Response)
	}

	return request
}
