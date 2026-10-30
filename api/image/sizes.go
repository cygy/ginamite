package image

// Sizes holds the different sizes of an image.
type Sizes struct {
	Size1x Property
	Size2x Property
}

// IsDefined returns true if the properties are defined.
func (s Sizes) IsDefined() bool {
	return s.Size1x.IsDefined() && s.Size2x.IsDefined()
}
