// Package shorthand provides ready-made handlers for the common pages.
package shorthand

import (
	"net/http"

	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/response"

	"github.com/gin-gonic/gin"
)

// APICallFunc is the API call validating the property of the user: it receives the query parameters "id" and "key"
// and returns the message to display, the result of the call, and whether it succeeded.
type APICallFunc func(context *gin.Context, id, key string) (string, api.Result, bool)

// PageConfigurationFunc configures the page before the API call.
type PageConfigurationFunc func(page *html.Page)

// RenderOverrideFunc renders the page instead of the default rendering.
type RenderOverrideFunc func(c *gin.Context, page *html.Page, message string, result api.Result, ok bool)

// BeforeRenderingFunc is called after the API call and before the page rendering.
type BeforeRenderingFunc func(c *gin.Context, page *html.Page, message string, result api.Result, ok bool)

// Validation validates a property, such as an email address, without authentication: the link sent to the user holds
// the query parameters "id" and "key".
type Validation struct {
	PageKey           string
	APICall           APICallFunc // Required.
	PageConfiguration PageConfigurationFunc
	BeforeRendering   BeforeRenderingFunc
	RenderOverride    RenderOverrideFunc
}

// NewValidation returns a Validation of the page targeted by the request. See context.SavePageKey.
func NewValidation(c *gin.Context) (v *Validation) {
	v = new(Validation)
	v.PageKey = context.GetPageKey(c)
	return
}

// Go executes the validation: it calls the API, then renders the page with the message returned by the call.
// If the call fails, nothing is left to do when its response was handled; else the error of the API is displayed
// by the page, or an error page is rendered if the API returned none.
func (v *Validation) Go(c *gin.Context) {
	page := html.Main.PageByKey(c, v.PageKey)
	if page == nil {
		response.NotFound(c)
		return
	}

	if v.PageConfiguration != nil {
		v.PageConfiguration(page)
	}

	message, result, ok := v.APICall(c, c.Query("id"), c.Query("key"))
	if !ok {
		if !result.Handled {
			v.renderFailure(c, page, result)
		}
		return
	}

	if v.BeforeRendering != nil {
		v.BeforeRendering(c, page, message, result, ok)
	}

	if v.RenderOverride != nil {
		v.RenderOverride(c, page, message, result, ok)
		return
	}

	page.SetInfo(html.Info{
		Message: message,
	})
	page.RenderOk(c)
}

// renderFailure renders the error of the API on the page, or a page matching the status of the response
// if the API returned no error.
func (v *Validation) renderFailure(c *gin.Context, page *html.Page, result api.Result) {
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
}
