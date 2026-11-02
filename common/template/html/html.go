// Package html loads the HTML templates of the application.
package html

import (
	"errors"
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
)

// LoadTemplates loads the HTML templates found in the sub-directories of path.
// See LoadTemplatesWithDelimiters.
func LoadTemplates(path string, e *gin.Engine) (*template.Template, error) {
	return LoadTemplatesWithDelimiters(path, "", "", e)
}

// LoadTemplatesWithDelimiters loads the HTML templates found in the sub-directories of path
// and, if e is not nil, into the Gin engine, with the helpers plus, minus, mult and div, and returns them.
// The delimiters are the default ones unless both left and right are set.
// It returns an error if a template can not be parsed.
func LoadTemplatesWithDelimiters(path, left, right string, e *gin.Engine) (*template.Template, error) {
	pattern := fmt.Sprintf("%s/**/*", path)
	customDelimiters := len(left) > 0 && len(right) > 0

	t := template.New("html").Funcs(funcMap())
	if customDelimiters {
		t = t.Delims(left, right)
	}

	t, err := t.ParseGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("unable to load the HTML templates of %s: %w", path, err)
	}

	// The templates are valid: the engine, which panics on an invalid one, parses them again and reloads them on
	// each request in debug mode.
	if e != nil {
		if customDelimiters {
			e.Delims(left, right)
		}
		e.SetFuncMap(funcMap())
		e.LoadHTMLGlob(pattern)
	}

	return t, nil
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
