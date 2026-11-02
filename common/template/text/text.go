// Package text loads the text templates of the application.
package text

import (
	"fmt"
	"text/template"
)

// LoadTemplates loads the text templates found in the sub-directories of path.
// See LoadTemplatesWithDelimiters.
func LoadTemplates(path string) (*template.Template, error) {
	return LoadTemplatesWithDelimiters(path, "", "")
}

// LoadTemplatesWithDelimiters loads the text templates found in the sub-directories of path and returns them.
// The delimiters are the default ones unless both left and right are set.
// It returns an error if a template can not be parsed.
func LoadTemplatesWithDelimiters(path, left, right string) (*template.Template, error) {
	t := template.New("text")
	if len(left) > 0 && len(right) > 0 {
		t = t.Delims(left, right)
	}

	t, err := t.ParseGlob(fmt.Sprintf("%s/**/*", path))
	if err != nil {
		return nil, fmt.Errorf("unable to load the text templates of %s: %w", path, err)
	}

	return t, nil
}
