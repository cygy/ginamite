package model

import "testing"

func TestNewPublicImage(t *testing.T) {
	got := NewPublicImage(testImage())
	want := PublicImage{
		Full:  PublicImageSize{Size1x: "full1", Size2x: "full2"},
		Small: PublicImageSize{Size1x: "small1", Size2x: "small2"},
		Thumb: PublicImageSize{Size1x: "thumb1", Size2x: "thumb2"},
	}
	if got != want {
		t.Errorf("got %+v", got)
	}
}

func TestNewPublicSummarizedImage(t *testing.T) {
	if got := NewPublicSummarizedImage(testImage()); got != (PublicSummarizedImage{Size1x: "small1", Size2x: "small2"}) {
		t.Errorf("got %+v", got)
	}
}

func TestNewPublicSummarizedImageFromPublicImage(t *testing.T) {
	got := NewPublicSummarizedImageFromPublicImage(NewPublicImage(testImage()))
	if got != (PublicSummarizedImage{Size1x: "small1", Size2x: "small2"}) {
		t.Errorf("got %+v", got)
	}
}
