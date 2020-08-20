package shorthand

import (
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"

	"github.com/gin-gonic/gin"
)

// APICallFunc : signature of an API call function.
type APICallFunc func(context *gin.Context, id, key string) (string, api.Result, bool)

// PageConfigurationFunc : function to configure the page if needed.
type PageConfigurationFunc func(page *html.Page)

// RenderOverrideFunc : function to override the page rendering.
type RenderOverrideFunc func(c *gin.Context, page *html.Page, message string, result api.Result, ok bool)

// BeforeRenderingFunc : function called after the API call and before the page rendering.
type BeforeRenderingFunc func(c *gin.Context, page *html.Page, message string, result api.Result, ok bool)

// Validation  : struct to validate a property without authentication.
type Validation struct {
	PageKey           string
	APICall           APICallFunc
	PageConfiguration PageConfigurationFunc
	BeforeRendering   BeforeRenderingFunc
	RenderOverride    RenderOverrideFunc
}

// NewValidation : returns a new Validation struct.
func NewValidation(c *gin.Context) (v *Validation) {
	v = new(Validation)
	v.PageKey = context.GetPageKey(c)
	return
}
