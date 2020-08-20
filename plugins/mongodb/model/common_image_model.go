package model

// Image : struct of a image.
type Image struct {
	Full  ImageSize `json:"full" bson:"full"`
	Small ImageSize `json:"small" bson:"small"`
	Thumb ImageSize `json:"thumbnail" bson:"thumb"`
}

// ImageSize : struct of the size of an image.
type ImageSize struct {
	Size1x ImageStorage `json:"size1x" bson:"size1x"`
	Size2x ImageStorage `json:"size2x" bson:"size2x"`
}

// ImageStorage : struct of the size of an image storage.
type ImageStorage struct {
	Path string `json:"path" bson:"path"`
	URL  string `json:"url" bson:"url"`
}
