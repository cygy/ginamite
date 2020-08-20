package web

import (
	"net/http"

	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/response"
	"github.com/gin-gonic/gin"
)

// HandleAPICallToPage : returns true if this function handles the result of the API call, else returns false.
func HandleAPICallToPage(c *gin.Context, page *html.Page, result api.Result, ok bool) bool {
	// The API call did not return an error, nothing to handle here.
	if ok {
		return false
	}

	if !result.Handled {
		// The error returned by the API is handled here.
		if !result.APIError.IsNil() {
			response.APIError(c, page, result.APIError)
		} else {
			// Handles the HTTP error here.
			switch result.StatusCode {
			case http.StatusNotFound:
				response.NotFound(c)
				break
			case http.StatusInternalServerError:
				response.InternalServerError(c)
				break
			case http.StatusUnauthorized:
				response.Unauthorized(c)
				break
			default:
			}
		}
	}

	return true
}

// HandleAPICallToPartialPage : returns true if this function handles the result of the API call, else returns false.
func HandleAPICallToPartialPage(c *gin.Context, result api.Result, ok bool) bool {
	if ok {
		return false
	}

	if !result.Handled {
		c.AbortWithStatusJSON(result.StatusCode, map[string]interface{}{"error": result.APIError})
	}

	return true
}
