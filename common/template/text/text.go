// Package text loads the text templates of the application.
package text

import (
	"fmt"
	"text/template"
)

// Main holds the text templates of the application, set by LoadTemplates.
var Main *template.Template

// LoadTemplates loads the text templates found in the sub-directories of path.
// See LoadTemplatesWithDelimiters.
func LoadTemplates(path string) {
	LoadTemplatesWithDelimiters(path, "", "")
}

// LoadTemplatesWithDelimiters loads the text templates found in the sub-directories of path into Main.
// The delimiters are the default ones unless both left and right are set.
// It panics if a template can not be parsed.
func LoadTemplatesWithDelimiters(path, left, right string) {
	t := template.New("text")
	if len(left) > 0 && len(right) > 0 {
		t = t.Delims(left, right)
	}

	Main = template.Must(t.ParseGlob(fmt.Sprintf("%s/**/*", path)))
}
