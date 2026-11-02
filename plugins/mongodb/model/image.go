package model

// Image is an image stored in several sizes.
type Image struct {
	Full  ImageSize `json:"full" bson:"full"`
	Small ImageSize `json:"small" bson:"small"`
	Thumb ImageSize `json:"thumbnail" bson:"thumb"`
}

// ImageSize is a size of an image, for the standard and the high density screens.
type ImageSize struct {
	Size1x ImageStorage `json:"size1x" bson:"size1x"`
	Size2x ImageStorage `json:"size2x" bson:"size2x"`
}

// ImageStorage is the location of a stored image file.
type ImageStorage struct {
	Path string `json:"path" bson:"path"`
	URL  string `json:"url" bson:"url"`
}
