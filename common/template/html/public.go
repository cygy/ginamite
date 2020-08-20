package html

import (
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
)

// LoadTemplates : loads the HTML templates.
func LoadTemplates(path string, e *gin.Engine) {
	LoadTemplatesWithDelimiters(path, "", "", e)
}

// LoadTemplatesWithDelimiters : loads the HTML templates.
func LoadTemplatesWithDelimiters(path, left, right string, e *gin.Engine) {
	pattern := fmt.Sprintf("%s/**/*", path)

	if e != nil {
		e.SetFuncMap(template.FuncMap{
			"plus": func(a, b int) int {
				return a + b
			},
			"minus": func(a, b int) int {
				return a - b
			},
			"mult": func(a, b int) int {
				return a * b
			},
			"div": func(a, b int) int {
				return a / b
			},
		})
		e.LoadHTMLGlob(pattern)
	}

	t := template.New("html")
	if len(left) > 0 && len(right) > 0 {
		t = t.Delims(left, right)
	}

	Main = template.Must(t.ParseGlob(pattern))
}
