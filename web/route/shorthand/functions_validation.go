package shorthand

import (
	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/response"

	"github.com/gin-gonic/gin"
)

// Go : execute the validation function.
func (v *Validation) Go(c *gin.Context) {
	page := html.Main.PageByKey(c, v.PageKey)

	if v.PageConfiguration != nil {
		v.PageConfiguration(page)
	}

	id := c.Query("id")
	key := c.Query("key")

	message, result, ok := v.APICall(c, id, key)
	if !ok {
		if !result.Handled {
			response.APIError(c, page, result.APIError)
		}
		return
	}

	if v.BeforeRendering != nil {
		v.BeforeRendering(c, page, message, result, ok)
	}

	if v.RenderOverride != nil {
		v.RenderOverride(c, page, message, result, ok)
	} else {
		page.SetInfo(html.Info{
			Message: message,
		})
		page.RenderOk(c)
	}
}
