package model

// PublicImage : struct of an image.
type PublicImage struct {
	Full  PublicImageSize `json:"full" bson:"full"`
	Small PublicImageSize `json:"small" bson:"small"`
	Thumb PublicImageSize `json:"thumbnail" bson:"thumb"`
}

// PublicImageSize : struct of the size of an image.
type PublicImageSize struct {
	Size1x string `json:"size1x" bson:"size1x"`
	Size2x string `json:"size2x" bson:"size2x"`
}

// NewPublicImage : returns a new 'PublicImage' struct.
func NewPublicImage(image Image) PublicImage {
	return PublicImage{
		Full: PublicImageSize{
			Size1x: image.Full.Size1x.URL,
			Size2x: image.Full.Size2x.URL,
		},
		Small: PublicImageSize{
			Size1x: image.Small.Size1x.URL,
			Size2x: image.Small.Size2x.URL,
		},
		Thumb: PublicImageSize{
			Size1x: image.Thumb.Size1x.URL,
			Size2x: image.Thumb.Size2x.URL,
		},
	}
}

// PublicSummarizedImage : struct of an image.
type PublicSummarizedImage PublicImageSize

// NewPublicSummarizedImage : returns a new 'PublicSummarizedImage' struct.
func NewPublicSummarizedImage(image Image) PublicSummarizedImage {
	return PublicSummarizedImage{
		Size1x: image.Small.Size1x.URL,
		Size2x: image.Small.Size2x.URL,
	}
}

// NewPublicSummarizedImageFromPublicImage : returns a new 'PublicSummarizedImage' struct.
func NewPublicSummarizedImageFromPublicImage(image PublicImage) PublicSummarizedImage {
	return NewPublicSummarizedImage(Image{
		Full: ImageSize{
			Size1x: ImageStorage{
				URL: image.Full.Size1x,
			},
			Size2x: ImageStorage{
				URL: image.Full.Size2x,
			},
		},
		Small: ImageSize{
			Size1x: ImageStorage{
				URL: image.Small.Size1x,
			},
			Size2x: ImageStorage{
				URL: image.Small.Size2x,
			},
		},
		Thumb: ImageSize{
			Size1x: ImageStorage{
				URL: image.Thumb.Size1x,
			},
			Size2x: ImageStorage{
				URL: image.Thumb.Size2x,
			},
		},
	})
}
