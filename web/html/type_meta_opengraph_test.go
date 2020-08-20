package html

import "testing"

func TestMetaOpenGraphCopy(t *testing.T) {
	original := OpenGraph{
		Title:       "original_title",
		Description: "original_description",
		Image:       "original_image",
		URL:         "original_url",
		SiteName:    "original_site_name",
		Type:        "original_type",
		Locale:      "original_locale",
	}

	copy := original

	if original.Title != copy.Title {
		t.Error("Copying a meta opengraph struct must not modify the Title property.")
	}
	if original.Description != copy.Description {
		t.Error("Copying a meta opengraph struct must not modify the Description property.")
	}
	if original.Image != copy.Image {
		t.Error("Copying a meta opengraph struct must not modify the Image property.")
	}
	if original.URL != copy.URL {
		t.Error("Copying a meta opengraph struct must not modify the URL property.")
	}
	if original.SiteName != copy.SiteName {
		t.Error("Copying a meta opengraph struct must not modify the SiteName property.")
	}
	if original.Type != copy.Type {
		t.Error("Copying a meta opengraph struct must not modify the Type property.")
	}
	if original.Locale != copy.Locale {
		t.Error("Copying a meta opengraph struct must not modify the Locale property.")
	}

	original.Title = original.Title + "_modified"
	original.Description = original.Description + "_modified"
	original.Image = original.Image + "_modified"
	original.URL = original.URL + "_modified"
	original.SiteName = original.SiteName + "_modified"
	original.Type = original.Type + "_modified"
	original.Locale = original.Locale + "_modified"

	if original.Title == copy.Title {
		t.Error("Modifying the Title property of a meta opengraph struct must not modify the Title property of the copied struct.")
	}
	if original.Description == copy.Description {
		t.Error("Modifying the Description property of a meta opengraph struct must not modify the Description property of the copied struct.")
	}
	if original.Image == copy.Image {
		t.Error("Modifying the Image property of a meta opengraph struct must not modify the Image property of the copied struct.")
	}
	if original.URL == copy.URL {
		t.Error("Modifying the URL property of a meta opengraph struct must not modify the URL property of the copied struct.")
	}
	if original.SiteName == copy.SiteName {
		t.Error("Modifying the SiteName property of a meta opengraph struct must not modify the SiteName property of the copied struct.")
	}
	if original.Type == copy.Type {
		t.Error("Modifying the Type property of a meta opengraph struct must not modify the Type property of the copied struct.")
	}
	if original.Locale == copy.Locale {
		t.Error("Modifying the Locale property of a meta opengraph struct must not modify the Locale property of the copied struct.")
	}
}

func TestMetaOpenGraphMerge(t *testing.T) {
	original := OpenGraph{
		Title:       "original_title",
		Description: "original_description",
		Image:       "original_image",
		URL:         "original_url",
		SiteName:    "original_site_name",
		Type:        "original_type",
		Locale:      "original_locale",
	}

	copy := OpenGraph{
		Title:       "copy_title",
		Description: "copy_description",
		Image:       "copy_image",
		URL:         "copy_url",
		SiteName:    "copy_site_name",
		Type:        "copy_type",
		Locale:      "copy_locale",
	}

	original.Merge(copy)

	if original.Title != copy.Title {
		t.Error("Merging a meta opengraph struct must modify the Title property.")
	}
	if original.Description != copy.Description {
		t.Error("Merging a meta opengraph struct must modify the Description property.")
	}
	if original.Image != copy.Image {
		t.Error("Merging a meta opengraph struct must modify the Image property.")
	}
	if original.URL != copy.URL {
		t.Error("Merging a meta opengraph struct must modify the URL property.")
	}
	if original.SiteName != copy.SiteName {
		t.Error("Merging a meta opengraph struct must modify the SiteName property.")
	}
	if original.Type != copy.Type {
		t.Error("Merging a meta opengraph struct must modify the Type property.")
	}
	if original.Locale != copy.Locale {
		t.Error("Merging a meta opengraph struct must modify the Locale property.")
	}

	original.Title = original.Title + "_modified"
	original.Description = original.Description + "_modified"
	original.Image = original.Image + "_modified"
	original.URL = original.URL + "_modified"
	original.SiteName = original.SiteName + "_modified"
	original.Type = original.Type + "_modified"
	original.Locale = original.Locale + "_modified"

	if original.Title == copy.Title {
		t.Error("Modifying the Title property of a meta opengraph struct must not modify the Title property of another struct.")
	}
	if original.Description == copy.Description {
		t.Error("Modifying the Description property of a meta opengraph struct must not modify the Description property of another struct.")
	}
	if original.Image == copy.Image {
		t.Error("Modifying the Image property of a meta opengraph struct must not modify the Image property of another struct.")
	}
	if original.URL == copy.URL {
		t.Error("Modifying the URL property of a meta opengraph struct must not modify the URL property of another struct.")
	}
	if original.SiteName == copy.SiteName {
		t.Error("Modifying the SiteName property of a meta opengraph struct must not modify the SiteName property of another struct.")
	}
	if original.Type == copy.Type {
		t.Error("Modifying the Type property of a meta opengraph struct must not modify the Type property of another struct.")
	}
	if original.Locale == copy.Locale {
		t.Error("Modifying the Locale property of a meta opengraph struct must not modify the Locale property of another struct.")
	}

	copy.Title = copy.Title + "_copied_and_modified"
	copy.Description = copy.Description + "_copied_and_modified"
	copy.Image = copy.Image + "_copied_and_modified"
	copy.URL = copy.URL + "_copied_and_modified"
	copy.SiteName = copy.SiteName + "_copied_and_modified"
	copy.Type = copy.Type + "_copied_and_modified"
	copy.Locale = copy.Locale + "_copied_and_modified"

	if original.Title == copy.Title {
		t.Error("Modifying the Title property of a meta opengraph struct must not modify the Title property of another struct.")
	}
	if original.Description == copy.Description {
		t.Error("Modifying the Description property of a meta opengraph struct must not modify the Description property of another struct.")
	}
	if original.Image == copy.Image {
		t.Error("Modifying the Image property of a meta opengraph struct must not modify the Image property of another struct.")
	}
	if original.URL == copy.URL {
		t.Error("Modifying the URL property of a meta opengraph struct must not modify the URL property of another struct.")
	}
	if original.SiteName == copy.SiteName {
		t.Error("Modifying the SiteName property of a meta opengraph struct must not modify the SiteName property of another struct.")
	}
	if original.Type == copy.Type {
		t.Error("Modifying the Type property of a meta opengraph struct must not modify the Type property of another struct.")
	}
	if original.Locale == copy.Locale {
		t.Error("Modifying the Locale property of a meta opengraph struct must not modify the Locale property of another struct.")
	}
}
