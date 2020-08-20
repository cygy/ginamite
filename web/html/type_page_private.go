package html

import (
	"bytes"
	"html/template"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Returns a localized string (used into the templates).
func (page *Page) rawString(key string) string {
	rawString := localization.Source(page.Locale, key)

	if len(rawString) == 0 {
		return key
	}

	return rawString
}

// Replaces the defined page variables into the string.
func (page *Page) replaceVars(entry string) string {
	return replaceVars(entry, page.Content)
}

// Replaces the defined page variables into the string.
func (page *Page) replaceURLVars(entry string, vars ...map[string]string) string {
	url := replaceVars(entry, nil)

	if len(vars) > 0 {
		for key, value := range vars[0] {
			url = strings.Replace(url, ":"+key, value, 1)
		}
	}

	return url
}

// Renders a HTML page with a HTTP code 500.
func (page *Page) render(c *gin.Context, httpCode int) {
	page.Process()
	c.HTML(httpCode, page.Template.Decorator, page)
	c.Abort()
}

// Replaces the defined page variables into the string.
func replaceVars(entry string, vars interface{}) string {
	var output bytes.Buffer

	t, err := template.New("_temp").Parse(entry)
	if err != nil {
		log.WithFields(logrus.Fields{
			"string": entry,
			"error":  err.Error(),
		}).Error("unable to parse the string")
		return ""
	}

	if err := t.Execute(&output, vars); err != nil {
		log.WithFields(logrus.Fields{
			"string": entry,
			"vars":   vars,
			"error":  err.Error(),
		}).Error("unable to execute the template")
		return ""
	}

	return output.String()
}

// Returns a date from a timzezone.
func timeInTimezone(t time.Time, tz string) (time.Time, error) {
	if locations == nil {
		locations = map[string]*time.Location{}
	}

	var err error
	location, ok := locations[tz]
	if !ok {
		location, err = time.LoadLocation(tz)
		if err == nil {
			locations[tz] = location
		}
	}

	if location != nil {
		t = t.In(location)
	}

	return t, err
}
