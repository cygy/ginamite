package html

// PageTemplate holds the names of the templates of a page.
type PageTemplate struct {
	Decorator string
	Main      string
}

// Merge overrides the templates with the ones defined in toMerge.
func (templates *PageTemplate) Merge(toMerge PageTemplate) {
	if len(toMerge.Decorator) > 0 {
		templates.Decorator = toMerge.Decorator
	}
	if len(toMerge.Main) > 0 {
		templates.Main = toMerge.Main
	}
}
