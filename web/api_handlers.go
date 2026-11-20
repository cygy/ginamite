package web

import (
	"net/http"

	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/response"

	"github.com/gin-gonic/gin"
)

// HandleAPICallToPage handles the result of an API call made for a page, and reports whether the caller must stop:
// it is false if the call succeeded.
//
// If the response was not already handled, the page displays the error of the API if there is one; else an error page
// matching the status of the response is rendered: not found, unauthorized or, for any other status, internal server error.
func HandleAPICallToPage(c *gin.Context, page *html.Page, result api.Result, ok bool) bool {
	if ok {
		return false
	}

	if result.Handled {
		return true
	}

	switch {
	case !result.APIError.IsNil():
		response.APIError(c, page, result.APIError)
	case result.StatusCode == http.StatusNotFound:
		response.NotFound(c)
	case result.StatusCode == http.StatusUnauthorized:
		response.Unauthorized(c)
	default:
		response.InternalServerError(c)
	}

	return true
}

// HandleAPICallToDataPage handles the result of an API call made for a page which can not be rendered without the data
// of the call (the page of an object, a list...), and reports whether the caller must stop: it is false if the call
// succeeded.
//
// Unlike [HandleAPICallToPage], the page never displays the error of the API, its template would run without data:
// if the response was not already handled, an error page matching the status of the response is rendered: not found
// for 404 and 400 (an invalid identifier), unauthorized for 401 and, for any other status, internal server error.
func HandleAPICallToDataPage(c *gin.Context, result api.Result, ok bool) bool {
	if ok {
		return false
	}

	if result.Handled {
		return true
	}

	switch result.StatusCode {
	case http.StatusNotFound, http.StatusBadRequest:
		response.NotFound(c)
	case http.StatusUnauthorized:
		response.Unauthorized(c)
	default:
		response.InternalServerError(c)
	}

	return true
}

// HandleAPICallToPartialPage handles the result of an API call made for a partial page, and reports whether the
// caller must stop: it is false if the call succeeded.
//
// If the response was not already handled, it is aborted with the status of the response, or 500 if the API did not
// respond, and a JSON body holding the error of the API.
func HandleAPICallToPartialPage(c *gin.Context, result api.Result, ok bool) bool {
	if ok {
		return false
	}

	if !result.Handled {
		statusCode := result.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}

		c.AbortWithStatusJSON(statusCode, map[string]any{"error": result.APIError})
	}

	return true
}
