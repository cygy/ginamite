package api

import (
	"github.com/cygy/ginamite/common/log"
	r "github.com/cygy/ginamite/common/request"
	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/cookie"

	"github.com/gin-gonic/gin"
	resty "github.com/go-resty/resty/v2"
)

// AddResponseHandlers : adds some response handlers to the current response handlers.
func (client *Client) AddResponseHandlers(handlers ...ResponseHandlerFunc) {
	client.responseHandlers = append(client.responseHandlers, handlers...)
}

// Call : direct call to the API.
func (client *Client) Call(c *gin.Context, req Request) (Result, bool) {
	// Add the locale.
	request := client.client.R().SetHeader(r.LocaleHeaderName, context.GetLocale(c))

	// Add the unique ID.
	requestID := c.Request.Header.Get(log.RequestIDHeaderName)
	if len(requestID) > 0 {
		request.SetHeader(log.RequestIDHeaderName, requestID)
	}

	// Add the user IP address
	userIPAddress := r.GetRealIPAddress(c)
	request.SetHeader(r.UserIPAddressHeaderName, userIPAddress)

	// Add the auth token.
	if req.Authenticated {
		if token, _ := cookie.GetValue(c); len(token) > 0 {
			request.SetAuthToken(token)
		}
	}

	// Add the content-type.
	contentType := req.ContentType
	if len(contentType) > 0 {
		request.SetHeader("Content-Type", contentType)
	}

	// Add the body.
	if req.Body != nil {
		request.SetBody(req.Body)
	}

	// Set up the response and the error.
	if req.Response != nil {
		request.SetResult(&req.Response)
	}

	var APIError struct {
		Error response.Error
	}
	request.SetError(&APIError)

	var res *resty.Response
	var requestError error

	switch req.Method {
	case GET:
		res, requestError = request.Get(req.Endpoint)
	case POST:
		res, requestError = request.Post(req.Endpoint)
	case PUT:
		res, requestError = request.Put(req.Endpoint)
	case PATCH:
		res, requestError = request.Patch(req.Endpoint)
	case DELETE:
		res, requestError = request.Delete(req.Endpoint)
	}

	statusCode := res.StatusCode()

	handled := false

	if req.UseDefaultResponseHandlers {
		for _, handler := range client.responseHandlers {
			if handler(c, statusCode, requestError, APIError.Error) {
				handled = true
				break
			}
		}
	}

	if !handled {
		for _, handler := range req.ResponseHandlers {
			if handler(c, statusCode, requestError, APIError.Error) {
				handled = true
				break
			}
		}
	}

	// Any API call can return a renewed auth token by the headers.
	// It must be saved in a cookie.
	renewedAuthToken := res.Header().Get(response.RenewedAuthTokenHeader)
	if len(renewedAuthToken) > 0 {
		cookie.Create(c, renewedAuthToken)
	}

	return Result{
		StatusCode:   statusCode,
		RequestError: requestError,
		APIError:     APIError.Error,
		Handled:      handled,
	}, (!handled && requestError == nil && APIError.Error.IsNil() && statusCode < 300)
}
