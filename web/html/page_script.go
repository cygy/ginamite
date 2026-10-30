package html

// Script is a script file of the page.
type Script struct {
	Src         string
	Integrity   string
	CrossOrigin string
	Defer       bool
	Async       bool
}
