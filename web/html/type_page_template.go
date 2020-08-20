package html

// PageTemplate struct containing the properties of templates.
type PageTemplate struct {
	Decorator string
	Main      string
}

// Merge merges two PageTemplate structs.
func (templates *PageTemplate) Merge(toMerge PageTemplate) {
	if len(toMerge.Decorator) > 0 {
		templates.Decorator = toMerge.Decorator
	}
	if len(toMerge.Main) > 0 {
		templates.Main = toMerge.Main
	}
}
