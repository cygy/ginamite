package model

// OwnImage : struct of an image.
type OwnImage struct {
	Full  OwnImageSize `json:"full" bson:"full"`
	Small OwnImageSize `json:"small" bson:"small"`
	Thumb OwnImageSize `json:"thumbnail" bson:"thumb"`
}

// OwnImageSize : struct of the size of an image.
type OwnImageSize struct {
	Size1x string `json:"size1x" bson:"size1x"`
	Size2x string `json:"size2x" bson:"size2x"`
}

// NewOwnImage : returns a new 'OwnImage' struct.
func NewOwnImage(image Image) OwnImage {
	return OwnImage{
		Full: OwnImageSize{
			Size1x: image.Full.Size1x.URL,
			Size2x: image.Full.Size2x.URL,
		},
		Small: OwnImageSize{
			Size1x: image.Small.Size1x.URL,
			Size2x: image.Small.Size2x.URL,
		},
		Thumb: OwnImageSize{
			Size1x: image.Thumb.Size1x.URL,
			Size2x: image.Thumb.Size2x.URL,
		},
	}
}
