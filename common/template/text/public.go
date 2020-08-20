package text

import (
	"fmt"
	"text/template"
)

// LoadTemplates : loads the text templates.
func LoadTemplates(path string) {
	LoadTemplatesWithDelimiters(path, "", "")
}

// LoadTemplatesWithDelimiters : loads the text templates.
func LoadTemplatesWithDelimiters(path, left, right string) {
	t := template.New("text")
	if len(left) > 0 && len(right) > 0 {
		t = t.Delims(left, right)
	}

	Main = template.Must(t.ParseGlob(fmt.Sprintf("%s/**/*", path)))
}
