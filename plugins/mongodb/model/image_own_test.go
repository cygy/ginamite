package model

import "testing"

// testImage returns an image whose URLs are the names of the sizes.
func testImage() Image {
	return Image{
		Full:  ImageSize{Size1x: ImageStorage{Path: "p", URL: "full1"}, Size2x: ImageStorage{URL: "full2"}},
		Small: ImageSize{Size1x: ImageStorage{URL: "small1"}, Size2x: ImageStorage{URL: "small2"}},
		Thumb: ImageSize{Size1x: ImageStorage{URL: "thumb1"}, Size2x: ImageStorage{URL: "thumb2"}},
	}
}

func TestNewOwnImage(t *testing.T) {
	got := NewOwnImage(testImage())
	want := OwnImage{
		Full:  OwnImageSize{Size1x: "full1", Size2x: "full2"},
		Small: OwnImageSize{Size1x: "small1", Size2x: "small2"},
		Thumb: OwnImageSize{Size1x: "thumb1", Size2x: "thumb2"},
	}
	if got != want {
		t.Errorf("got %+v", got)
	}
}
