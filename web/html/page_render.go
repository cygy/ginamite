package html

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Process replaces the variables of the page into its properties, then executes its main template: the result is
// included to the decorator template by IncludeMainTemplate.
// A template error is logged, and leaves the main template empty.
func (page *Page) Process() {
	page.Content.Title = page.replaceVars(page.Content.Title)

	page.Meta.Title = page.replaceVars(page.Meta.Title)
	page.Meta.Description = page.replaceVars(page.Meta.Description)
	page.Meta.Twitter.ImageAlt = page.replaceVars(page.Meta.Twitter.ImageAlt)

	for i, keyword := range page.Meta.Keywords {
		page.Meta.Keywords[i] = page.replaceVars(keyword)
	}

	for i, hrefLang := range page.Meta.HrefLang {
		page.Meta.HrefLang[i].Href = page.replaceVars(hrefLang.Href)
	}

	// Choose the right canonical URL.
	page.Meta.setUpCanonicalURL(page.Locale)

	// Set the OpenGraph tags up.
	page.Meta.OpenGraph.URL = page.Meta.CanonicalURL
	page.Meta.OpenGraph.SiteName = page.replaceVars(page.Meta.OpenGraph.SiteName)

	if len(page.Meta.OpenGraph.Title) == 0 {
		page.Meta.OpenGraph.Title = page.Meta.Title
	} else {
		page.Meta.OpenGraph.Title = page.replaceVars(page.Meta.OpenGraph.Title)
	}

	if len(page.Meta.OpenGraph.Description) == 0 {
		page.Meta.OpenGraph.Description = page.Meta.Description
	} else {
		page.Meta.OpenGraph.Description = page.replaceVars(page.Meta.OpenGraph.Description)
	}

	// Parse the main template: it will be included to the decorator template later.
	var output bytes.Buffer
	if err := page.executeMainTemplate(&output); err != nil {
		log.WithFields(logrus.Fields{
			"template": page.Template.Main,
			"page":     page.Key,
			"error":    err.Error(),
		}).Error("unable to execute the template")
	} else {
		page.Content.parsedMainTemplate = output.String()
	}
}

// executeMainTemplate executes the main template of the page. It fails if the settings have no templates.
func (page *Page) executeMainTemplate(output *bytes.Buffer) error {
	templates := page.settings().Templates
	if templates == nil {
		return errors.New("no templates")
	}

	return templates.ExecuteTemplate(output, page.Template.Main, page)
}

// RenderOk renders the page with the HTTP code 200.
func (page *Page) RenderOk(c *gin.Context) {
	page.render(c, http.StatusOK)
}

// RenderError renders the page with the HTTP code 500.
func (page *Page) RenderError(c *gin.Context) {
	page.render(c, http.StatusInternalServerError)
}

// RenderUnauthorized renders the page with the HTTP code 401.
func (page *Page) RenderUnauthorized(c *gin.Context) {
	page.render(c, http.StatusUnauthorized)
}

// RenderNotFound renders the page with the HTTP code 404.
func (page *Page) RenderNotFound(c *gin.Context) {
	page.render(c, http.StatusNotFound)
}

// SetError defines the error displayed by the page.
func (page *Page) SetError(err response.Error) {
	page.Content.Error.Message = err.Message
	page.Content.Error.Reason = err.Reason
	page.Content.Error.Recovery = err.Recovery
}

// SetInfo defines the information displayed by the page.
func (page *Page) SetInfo(info Info) {
	page.Content.Info.Title = info.Title
	page.Content.Info.Message = info.Message
}

// render processes the page, sends it with the HTTP code and aborts the following handlers.
func (page *Page) render(c *gin.Context, httpCode int) {
	page.Process()
	c.HTML(httpCode, page.Template.Decorator, page)
	c.Abort()
}
