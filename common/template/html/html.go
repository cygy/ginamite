// Package html loads the HTML templates of the application.
package html

import (
	"errors"
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
)

// Main holds the HTML templates of the application, set by LoadTemplates.
var Main *template.Template

// LoadTemplates loads the HTML templates found in the sub-directories of path.
// See LoadTemplatesWithDelimiters.
func LoadTemplates(path string, e *gin.Engine) {
	LoadTemplatesWithDelimiters(path, "", "", e)
}

// LoadTemplatesWithDelimiters loads the HTML templates found in the sub-directories of path
// into Main and, if e is not nil, into the Gin engine, with the helpers plus, minus, mult and div.
// The delimiters are the default ones unless both left and right are set.
// It panics if a template can not be parsed.
func LoadTemplatesWithDelimiters(path, left, right string, e *gin.Engine) {
	pattern := fmt.Sprintf("%s/**/*", path)
	customDelimiters := len(left) > 0 && len(right) > 0

	if e != nil {
		if customDelimiters {
			e.Delims(left, right)
		}
		e.SetFuncMap(funcMap())
		e.LoadHTMLGlob(pattern)
	}

	t := template.New("html").Funcs(funcMap())
	if customDelimiters {
		t = t.Delims(left, right)
	}

	Main = template.Must(t.ParseGlob(pattern))
}

// funcMap returns the helpers available in the templates.
// Division by zero is an error of the template, not a panic.
func funcMap() template.FuncMap {
	return template.FuncMap{
		"plus":  func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
		"mult":  func(a, b int) int { return a * b },
		"div": func(a, b int) (int, error) {
			if b == 0 {
				return 0, errors.New("division by zero")
			}

			return a / b, nil
		},
	}
}
