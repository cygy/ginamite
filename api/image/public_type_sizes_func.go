package image

// IsDefined : returns true if the properties are defined.
func (s Sizes) IsDefined() bool {
	return s.Size1x.IsDefined() && s.Size2x.IsDefined()
}
