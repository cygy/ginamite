package html

import "strings"

// Meta struct of a HTML meta data.
type Meta struct {
	Charset      string
	Title        string
	Description  string
	Keywords     []string
	CanonicalURL string `yaml:"-"` // Defined by the constructor
	Icons        []Icon
	ImageSource  string `yaml:"image_source"`
	Robots       string
	ViewPort     string
	HrefLang     []HrefLang `yaml:"-"` // Defined by the constructor
	Alternates   []Alternate
	Stylesheets  []Stylesheet
	OpenGraph    OpenGraph
	Twitter      Twitter
}

// Copy copied a struct into an other.
func (meta *Meta) Copy() (copied *Meta) {
	copied = &Meta{}
	copied.Charset = meta.Charset
	copied.Title = meta.Title
	copied.Description = meta.Description
	copied.CanonicalURL = meta.CanonicalURL
	copied.ImageSource = meta.ImageSource
	copied.Robots = meta.Robots
	copied.ViewPort = meta.ViewPort
	copied.OpenGraph = meta.OpenGraph
	copied.Twitter = meta.Twitter

	copied.Keywords = make([]string, len(meta.Keywords))
	copy(copied.Keywords, meta.Keywords)

	copied.Icons = make([]Icon, len(meta.Icons))
	for i, icon := range meta.Icons {
		copied.Icons[i] = icon
	}

	copied.HrefLang = make([]HrefLang, len(meta.HrefLang))
	for i, hrefLang := range meta.HrefLang {
		copied.HrefLang[i] = hrefLang
	}

	copied.Alternates = make([]Alternate, len(meta.Alternates))
	for i, alternate := range meta.Alternates {
		copied.Alternates[i] = alternate
	}

	copied.Stylesheets = make([]Stylesheet, len(meta.Stylesheets))
	for i, stylesheet := range meta.Stylesheets {
		copied.Stylesheets[i] = stylesheet
	}

	return
}

// Merge merges two Meta structs.
func (meta *Meta) Merge(toMerge Meta) {
	if len(toMerge.Charset) > 0 {
		meta.Charset = toMerge.Charset
	}
	if len(toMerge.Title) > 0 {
		meta.Title = toMerge.Title
	}
	if len(toMerge.Description) > 0 {
		meta.Description = toMerge.Description
	}
	if len(toMerge.CanonicalURL) > 0 {
		meta.CanonicalURL = toMerge.CanonicalURL
	}
	if len(toMerge.ImageSource) > 0 {
		meta.ImageSource = toMerge.ImageSource
	}
	if len(toMerge.Robots) > 0 {
		meta.Robots = toMerge.Robots
	}
	if len(toMerge.ViewPort) > 0 {
		meta.ViewPort = toMerge.ViewPort
	}
	meta.OpenGraph.Merge(toMerge.OpenGraph)
	meta.Twitter.Merge(toMerge.Twitter)
	if toMerge.Keywords != nil {
		meta.Keywords = append(meta.Keywords, toMerge.Keywords...)
	}
	if toMerge.Icons != nil {
		meta.Icons = append(meta.Icons, toMerge.Icons...)
	}
	if toMerge.HrefLang != nil {
		meta.HrefLang = append(meta.HrefLang, toMerge.HrefLang...)
	}
	if toMerge.Alternates != nil {
		meta.Alternates = append(meta.Alternates, toMerge.Alternates...)
	}
	if toMerge.Stylesheets != nil {
		meta.Stylesheets = append(meta.Stylesheets, toMerge.Stylesheets...)
	}
}

// JoinedKeywords returns a string containing the keywords separated by a comma.
func (meta *Meta) JoinedKeywords() string {
	return strings.Join(meta.Keywords, ", ")
}
