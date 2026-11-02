package image

// Versions holds the different versions of an image.
type Versions struct {
	Full  Sizes
	Small Sizes
	Thumb Sizes
}

// All returns all the properties of the image.
func (v Versions) All() []Property {
	properties := []Property{v.Full.Size1x, v.Full.Size2x}

	if v.Small.IsDefined() {
		properties = append(properties, v.Small.Size1x, v.Small.Size2x)
	}

	if v.Thumb.IsDefined() {
		properties = append(properties, v.Thumb.Size1x, v.Thumb.Size2x)
	}

	return properties
}
