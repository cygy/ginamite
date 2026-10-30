package html

// Twitter is the Twitter metadata of a page.
type Twitter struct {
	Card     string
	ImageAlt string `yaml:"image_alt"`
	Site     string // Twitter account i.e. '@cyril'
	Creator  string // Twitter account i.e. '@cyril'
	// The other properties are given by the opengraph properties.
}

// Merge overrides the properties with the ones defined in toMerge.
func (twitter *Twitter) Merge(toMerge Twitter) {
	if len(toMerge.Card) > 0 {
		twitter.Card = toMerge.Card
	}
	if len(toMerge.ImageAlt) > 0 {
		twitter.ImageAlt = toMerge.ImageAlt
	}
	if len(toMerge.Site) > 0 {
		twitter.Site = toMerge.Site
	}
	if len(toMerge.Creator) > 0 {
		twitter.Creator = toMerge.Creator
	}
}
