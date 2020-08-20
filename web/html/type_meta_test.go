package html

import (
	"reflect"
	"testing"
)

func TestMetaCopy(t *testing.T) {
	original := Meta{
		Charset:      "original_charset",
		Title:        "original_title",
		Description:  "original_description",
		Keywords:     []string{"keyword1", "keyword2"},
		CanonicalURL: "original_canonical_url",
		Icons: []Icon{Icon{
			Rel:    "original_rel",
			Href:   "original_href",
			Type:   "original_type",
			Color:  "original_color",
			Width:  1,
			Height: 1,
		}},
		ImageSource: "original_image_source",
		Robots:      "original_robots",
		ViewPort:    "original_viewport",
		HrefLang: []HrefLang{HrefLang{
			Href:   "original_href",
			Locale: "original_locale",
		}},
		Alternates: []Alternate{Alternate{
			Href:  "original_href",
			Type:  "original_type",
			Title: "original_title",
		}},
		Stylesheets: []Stylesheet{Stylesheet{
			Href:  "original_href",
			Media: "original_media",
		}},
		OpenGraph: OpenGraph{
			Title:       "original_title",
			Description: "original_description",
			Image:       "original_image",
			URL:         "original_url",
			SiteName:    "original_site_name",
			Locale:      "original_locale",
			Type:        "original_type",
		},
		Twitter: Twitter{
			Card:     "original_card",
			ImageAlt: "original_image_alt",
			Site:     "original_site",
		},
	}

	copy := original.Copy()

	if original.Charset != copy.Charset {
		t.Error("Copying a meta struct must not modify the Charset property.")
	}
	if original.Title != copy.Title {
		t.Error("Copying a meta struct must not modify the Title property.")
	}
	if original.Description != copy.Description {
		t.Error("Copying a meta struct must not modify the Description property.")
	}
	if original.CanonicalURL != copy.CanonicalURL {
		t.Error("Copying a meta struct must not modify the CanonicalURL property.")
	}
	if original.ImageSource != copy.ImageSource {
		t.Error("Copying a meta struct must not modify the ImageSource property.")
	}
	if original.Robots != copy.Robots {
		t.Error("Copying a meta struct must not modify the Robots property.")
	}
	if original.ViewPort != copy.ViewPort {
		t.Error("Copying a meta struct must not modify the ViewPort property.")
	}
	if !reflect.DeepEqual(original.Keywords, copy.Keywords) {
		t.Error("Copying a meta struct must not modify the Keywords property.")
	}
	if !reflect.DeepEqual(original.Icons, copy.Icons) {
		t.Error("Copying a meta struct must not modify the Icons property.")
	}
	if !reflect.DeepEqual(original.HrefLang, copy.HrefLang) {
		t.Error("Copying a meta struct must not modify the HrefLang property.")
	}
	if !reflect.DeepEqual(original.Alternates, copy.Alternates) {
		t.Error("Copying a meta struct must not modify the Alternates property.")
	}
	if !reflect.DeepEqual(original.Stylesheets, copy.Stylesheets) {
		t.Error("Copying a meta struct must not modify the Stylesheets property.")
	}
	if !original.OpenGraph.Equals(&copy.OpenGraph) {
		t.Error("Copying a meta struct must not modify the OpenGraph property.")
	}
	if original.Twitter != copy.Twitter {
		t.Error("Copying a meta struct must not modify the Twitter property.")
	}

	original.Charset = original.Charset + "_modified"
	original.Title = original.Title + "_modified"
	original.Description = original.Description + "_modified"
	original.CanonicalURL = original.CanonicalURL + "_modified"
	original.ImageSource = original.ImageSource + "_modified"
	original.Robots = original.Robots + "_modified"
	original.ViewPort = original.ViewPort + "_modified"
	for i, keyword := range original.Keywords {
		original.Keywords[i] = keyword + "_modified"
	}
	for i, icon := range original.Icons {
		original.Icons[i] = Icon{
			Rel:    icon.Rel + "_modified",
			Href:   icon.Href + "_modified",
			Type:   icon.Type + "_modified",
			Color:  icon.Color + "_modified",
			Width:  icon.Width + 1,
			Height: icon.Height + 1,
		}
	}
	for i, hreflang := range original.HrefLang {
		original.HrefLang[i] = HrefLang{
			Href:   hreflang.Href + "_modified",
			Locale: hreflang.Locale + "_modified",
		}
	}
	for i, alternate := range original.Alternates {
		original.Alternates[i] = Alternate{
			Href:  alternate.Href + "_modified",
			Title: alternate.Title + "_modified",
			Type:  alternate.Type + "_modified",
		}
	}
	for i, stylesheet := range original.Stylesheets {
		original.Stylesheets[i] = Stylesheet{
			Href:  stylesheet.Href + "_modified",
			Media: stylesheet.Media + "_modified",
		}
	}
	original.OpenGraph = OpenGraph{
		Title:       original.OpenGraph.Title + "_modified",
		Description: original.OpenGraph.Description + "_modified",
		Image:       original.OpenGraph.Image + "_modified",
		URL:         original.OpenGraph.URL + "_modified",
		SiteName:    original.OpenGraph.SiteName + "_modified",
		Locale:      original.OpenGraph.Locale + "_modified",
		Type:        original.OpenGraph.Type + "_modified",
	}
	original.Twitter = Twitter{
		Card:     original.Twitter.ImageAlt + "_modified",
		ImageAlt: original.Twitter.ImageAlt + "_modified",
		Site:     original.Twitter.Site + "_modified",
	}

	if original.Charset == copy.Charset {
		t.Error("Modifying the Charset property of a meta struct must not modify the Charset property of the copied struct.")
	}
	if original.Title == copy.Title {
		t.Error("Modifying the Title property of a meta struct must not modify the Title property of the copied struct.")
	}
	if original.Description == copy.Description {
		t.Error("Modifying the Description property of a meta struct must not modify the Description property of the copied struct.")
	}
	if original.CanonicalURL == copy.CanonicalURL {
		t.Error("Modifying the CanonicalURL property of a meta struct must not modify the CanonicalURL property of the copied struct.")
	}
	if original.ImageSource == copy.ImageSource {
		t.Error("Modifying the ImageSource property of a meta struct must not modify the ImageSource property of the copied struct.")
	}
	if original.Robots == copy.Robots {
		t.Error("Modifying the Robots property of a meta struct must not modify the Robots property of the copied struct.")
	}
	if original.ViewPort == copy.ViewPort {
		t.Error("Modifying the ViewPort property of a meta struct must not modify the ViewPort property of the copied struct.")
	}
	if reflect.DeepEqual(original.Keywords, copy.Keywords) {
		t.Error("Modifying the Keywords property of a meta struct must not modify the Keywords property of the copied struct.")
	}
	if reflect.DeepEqual(original.Icons, copy.Icons) {
		t.Error("Modifying the Icons property of a meta struct must not modify the Icons property of the copied struct.")
	}
	if reflect.DeepEqual(original.HrefLang, copy.HrefLang) {
		t.Error("Modifying the HrefLang property of a meta struct must not modify the HrefLang property of the copied struct.")
	}
	if reflect.DeepEqual(original.Alternates, copy.Alternates) {
		t.Error("Modifying the Alternates property of a meta struct must not modify the Alternates property of the copied struct.")
	}
	if reflect.DeepEqual(original.Stylesheets, copy.Stylesheets) {
		t.Error("Modifying the Stylesheets property of a meta struct must not modify the Stylesheets property of the copied struct.")
	}
	if !original.OpenGraph.Equals(&copy.OpenGraph) {
		t.Error("Modifying the OpenGraph property of a meta struct must not modify the OpenGraph property of the copied struct.")
	}
	if original.Twitter == copy.Twitter {
		t.Error("Modifying the Twitter property of a meta struct must not modify the Twitter property of the copied struct.")
	}
}

func TestMetaMerge(t *testing.T) {
	original := Meta{
		Charset:      "original_charset",
		Title:        "original_title",
		Description:  "original_description",
		Keywords:     []string{"keyword1", "keyword2"},
		CanonicalURL: "original_canonical_url",
		Icons: []Icon{Icon{
			Rel:    "original_rel",
			Href:   "original_href",
			Type:   "original_type",
			Color:  "original_color",
			Width:  1,
			Height: 1,
		}},
		ImageSource: "original_image_source",
		Robots:      "original_robots",
		ViewPort:    "original_viewport",
		HrefLang: []HrefLang{HrefLang{
			Href:   "original_href",
			Locale: "original_locale",
		}},
		Alternates: []Alternate{Alternate{
			Href:  "original_href",
			Type:  "original_type",
			Title: "original_title",
		}},
		Stylesheets: []Stylesheet{Stylesheet{
			Href:  "original_href",
			Media: "original_media",
		}},
		OpenGraph: OpenGraph{
			Title:       "original_title",
			Description: "original_description",
			Image:       "original_image",
			URL:         "original_url",
			SiteName:    "original_site_name",
			Locale:      "original_locale",
			Type:        "original_type",
		},
		Twitter: Twitter{
			Card:     "original_card",
			ImageAlt: "original_image_alt",
			Site:     "original_site",
		},
	}

	copy := Meta{
		Charset:      "copy_charset",
		Title:        "copy_title",
		Description:  "copy_description",
		Keywords:     []string{"keyword3", "keyword4"},
		CanonicalURL: "copy_canonical_url",
		Icons: []Icon{Icon{
			Rel:    "copy_rel",
			Href:   "copy_href",
			Type:   "copy_type",
			Color:  "copy_color",
			Width:  2,
			Height: 2,
		}},
		ImageSource: "copy_image_source",
		Robots:      "copy_robots",
		ViewPort:    "copy_viewport",
		HrefLang: []HrefLang{HrefLang{
			Href:   "copy_href",
			Locale: "copy_locale",
		}},
		Alternates: []Alternate{Alternate{
			Href:  "copy_href",
			Type:  "copy_type",
			Title: "copy_title",
		}},
		Stylesheets: []Stylesheet{Stylesheet{
			Href:  "copy_href",
			Media: "copy_media",
		}},
		OpenGraph: OpenGraph{
			Title:       "copy_title",
			Description: "copy_description",
			Image:       "copy_image",
			URL:         "copy_url",
			SiteName:    "copy_site_name",
			Locale:      "copy_locale",
			Type:        "copy_type",
		},
		Twitter: Twitter{
			Card:     "copy_card",
			ImageAlt: "copy_image_alt",
			Site:     "copy_site",
		},
	}

	originalKeywords := original.Keywords
	originalIcons := original.Icons
	originalHrefLangs := original.HrefLang
	originalAlternates := original.Alternates
	originalStylesheets := original.Stylesheets

	original.Merge(copy)

	if original.Charset != copy.Charset {
		t.Error("Merging a meta struct must modify the Charset property.")
	}
	if original.Title != copy.Title {
		t.Error("Merging a meta struct must modify the Title property.")
	}
	if original.Description != copy.Description {
		t.Error("Merging a meta struct must modify the Description property.")
	}
	if original.CanonicalURL != copy.CanonicalURL {
		t.Error("Merging a meta struct must modify the CanonicalURL property.")
	}
	if original.ImageSource != copy.ImageSource {
		t.Error("Merging a meta struct must modify the ImageSource property.")
	}
	if original.Robots != copy.Robots {
		t.Error("Merging a meta struct must modify the Robots property.")
	}
	if original.ViewPort != copy.ViewPort {
		t.Error("Merging a meta struct must modify the ViewPort property.")
	}
	if !original.OpenGraph.Equals(&copy.OpenGraph) {
		t.Error("Merging a meta struct must modify the OpenGraph property.")
	}
	if original.Twitter != copy.Twitter {
		t.Error("Merging a meta struct must modify the Twitter property.")
	}
	keywords := append(originalKeywords, copy.Keywords...)
	if !reflect.DeepEqual(original.Keywords, keywords) {
		t.Error("Merging a meta struct must merge all the keywords.")
	}
	icons := append(originalIcons, copy.Icons...)
	if !reflect.DeepEqual(original.Icons, icons) {
		t.Error("Merging a meta struct must merge all the icons.")
	}
	hrefLangs := append(originalHrefLangs, copy.HrefLang...)
	if !reflect.DeepEqual(original.HrefLang, hrefLangs) {
		t.Error("Merging a meta struct must merge all the href langs.")
	}
	alternates := append(originalAlternates, copy.Alternates...)
	if !reflect.DeepEqual(original.Alternates, alternates) {
		t.Error("Merging a meta struct must merge all the alternates.")
	}
	stylesheets := append(originalStylesheets, copy.Stylesheets...)
	if !reflect.DeepEqual(original.Stylesheets, stylesheets) {
		t.Error("Merging a meta struct must merge all the stylesheets.")
	}

	original.Charset = original.Charset + "_modified"
	original.Title = original.Title + "_modified"
	original.Description = original.Description + "_modified"
	original.CanonicalURL = original.CanonicalURL + "_modified"
	original.ImageSource = original.ImageSource + "_modified"
	original.Robots = original.Robots + "_modified"
	original.ViewPort = original.ViewPort + "_modified"
	for i, keyword := range original.Keywords {
		original.Keywords[i] = keyword + "_modified"
	}
	for i, icon := range original.Icons {
		original.Icons[i] = Icon{
			Rel:    icon.Rel + "_modified",
			Href:   icon.Href + "_modified",
			Type:   icon.Type + "_modified",
			Color:  icon.Color + "_modified",
			Width:  icon.Width + 1,
			Height: icon.Height + 1,
		}
	}
	for i, hreflang := range original.HrefLang {
		original.HrefLang[i] = HrefLang{
			Href:   hreflang.Href + "_modified",
			Locale: hreflang.Locale + "_modified",
		}
	}
	for i, alternate := range original.Alternates {
		original.Alternates[i] = Alternate{
			Href:  alternate.Href + "_modified",
			Title: alternate.Title + "_modified",
			Type:  alternate.Type + "_modified",
		}
	}
	for i, stylesheet := range original.Stylesheets {
		original.Stylesheets[i] = Stylesheet{
			Href:  stylesheet.Href + "_modified",
			Media: stylesheet.Media + "_modified",
		}
	}
	original.OpenGraph = OpenGraph{
		Title:       original.OpenGraph.Title + "_modified",
		Description: original.OpenGraph.Description + "_modified",
		Image:       original.OpenGraph.Image + "_modified",
		URL:         original.OpenGraph.URL + "_modified",
		SiteName:    original.OpenGraph.SiteName + "_modified",
		Locale:      original.OpenGraph.Locale + "_modified",
		Type:        original.OpenGraph.Type + "_modified",
	}
	original.Twitter = Twitter{
		Card:     original.Twitter.ImageAlt + "_modified",
		ImageAlt: original.Twitter.ImageAlt + "_modified",
		Site:     original.Twitter.Site + "_modified",
	}

	if original.Charset == copy.Charset {
		t.Error("Modifying the Charset property of a meta struct must not modify the Charset property of another struct.")
	}
	if original.Title == copy.Title {
		t.Error("Modifying the Title property of a meta struct must not modify the Title property of another struct.")
	}
	if original.Description == copy.Description {
		t.Error("Modifying the Description property of a meta struct must not modify the Description property of another struct.")
	}
	if original.CanonicalURL == copy.CanonicalURL {
		t.Error("Modifying the CanonicalURL property of a meta struct must not modify the CanonicalURL property of another struct.")
	}
	if original.ImageSource == copy.ImageSource {
		t.Error("Modifying the ImageSource property of a meta struct must not modify the ImageSource property of another struct.")
	}
	if original.Robots == copy.Robots {
		t.Error("Modifying the Robots property of a meta struct must not modify the Robots property of another struct.")
	}
	if original.ViewPort == copy.ViewPort {
		t.Error("Modifying the ViewPort property of a meta struct must not modify the ViewPort property of another struct.")
	}
	if !original.OpenGraph.Equals(&copy.OpenGraph) {
		t.Error("Modifying the OpenGraph property of a meta struct must not modify the OpenGraph property of another struct.")
	}
	if original.Twitter == copy.Twitter {
		t.Error("Modifying the Twitter property of a meta struct must not modify the Twitter property of another struct.")
	}
	if reflect.DeepEqual(original.Icons, copy.Icons) {
		t.Error("Modifying the Icons property of a meta struct must not modify the Icons property of another struct.")
	}
	if reflect.DeepEqual(original.Keywords, copy.Keywords) {
		t.Error("Modifying the Keywords property of a meta struct must not modify the Keywords property of another struct.")
	}
	if reflect.DeepEqual(original.HrefLang, copy.HrefLang) {
		t.Error("Modifying the HrefLang property of a meta struct must not modify the HrefLang property of another struct.")
	}
	if reflect.DeepEqual(original.Alternates, copy.Alternates) {
		t.Error("Modifying the Alternates property of a meta struct must not modify the Alternates property of another struct.")
	}
	if reflect.DeepEqual(original.Stylesheets, copy.Stylesheets) {
		t.Error("Modifying the Stylesheets property of a meta struct must not modify the Stylesheets property of another struct.")
	}

	copy.Charset = copy.Charset + "_copied_and_modified"
	copy.Title = copy.Title + "_copied_and_modified"
	copy.Description = copy.Description + "_copied_and_modified"
	copy.CanonicalURL = copy.CanonicalURL + "_copied_and_modified"
	copy.ImageSource = copy.ImageSource + "_copied_and_modified"
	copy.Robots = copy.Robots + "_copied_and_modified"
	copy.ViewPort = copy.ViewPort + "_copied_and_modified"
	for i, keyword := range copy.Keywords {
		copy.Keywords[i] = keyword + "_copied_and_modified"
	}
	for i, icon := range copy.Icons {
		copy.Icons[i] = Icon{
			Rel:    icon.Rel + "_copied_and_modified",
			Href:   icon.Href + "_copied_and_modified",
			Type:   icon.Type + "_copied_and_modified",
			Color:  icon.Color + "_copied_and_modified",
			Width:  icon.Width + 2,
			Height: icon.Height + 2,
		}
	}
	for i, hreflang := range copy.HrefLang {
		copy.HrefLang[i] = HrefLang{
			Href:   hreflang.Href + "_copied_and_modified",
			Locale: hreflang.Locale + "_copied_and_modified",
		}
	}
	for i, alternate := range copy.Alternates {
		copy.Alternates[i] = Alternate{
			Href:  alternate.Href + "_copied_and_modified",
			Title: alternate.Title + "_copied_and_modified",
			Type:  alternate.Type + "_copied_and_modified",
		}
	}
	for i, stylesheet := range copy.Stylesheets {
		copy.Stylesheets[i] = Stylesheet{
			Href:  stylesheet.Href + "_copied_and_modified",
			Media: stylesheet.Media + "_copied_and_modified",
		}
	}
	copy.OpenGraph = OpenGraph{
		Title:       copy.OpenGraph.Title + "_copied_and_modified",
		Description: copy.OpenGraph.Description + "_copied_and_modified",
		Image:       copy.OpenGraph.Image + "_copied_and_modified",
		URL:         copy.OpenGraph.URL + "_copied_and_modified",
		SiteName:    copy.OpenGraph.SiteName + "_copied_and_modified",
		Locale:      original.OpenGraph.Locale + "_copied_and_modified",
		Type:        original.OpenGraph.Type + "_copied_and_modified",
	}
	copy.Twitter = Twitter{
		Card:     copy.Twitter.ImageAlt + "_copied_and_modified",
		ImageAlt: copy.Twitter.ImageAlt + "_copied_and_modified",
		Site:     copy.Twitter.Site + "_copied_and_modified",
	}

	if original.Charset == copy.Charset {
		t.Error("Modifying the Charset property of a meta struct must not modify the Charset property of another struct.")
	}
	if original.Title == copy.Title {
		t.Error("Modifying the Title property of a meta struct must not modify the Title property of another struct.")
	}
	if original.Description == copy.Description {
		t.Error("Modifying the Description property of a meta struct must not modify the Description property of another struct.")
	}
	if original.CanonicalURL == copy.CanonicalURL {
		t.Error("Modifying the CanonicalURL property of a meta struct must not modify the CanonicalURL property of another struct.")
	}
	if original.ImageSource == copy.ImageSource {
		t.Error("Modifying the ImageSource property of a meta struct must not modify the ImageSource property of another struct.")
	}
	if original.Robots == copy.Robots {
		t.Error("Modifying the Robots property of a meta struct must not modify the Robots property of another struct.")
	}
	if original.ViewPort == copy.ViewPort {
		t.Error("Modifying the ViewPort property of a meta struct must not modify the ViewPort property of another struct.")
	}
	if !original.OpenGraph.Equals(&copy.OpenGraph) {
		t.Error("Modifying the OpenGraph property of a meta struct must not modify the OpenGraph property of another struct.")
	}
	if original.Twitter == copy.Twitter {
		t.Error("Modifying the Twitter property of a meta struct must not modify the Twitter property of another struct.")
	}
	if reflect.DeepEqual(original.Keywords, copy.Keywords) {
		t.Error("Modifying the Keywords property of a meta struct must not modify the Keywords property of another struct.")
	}
	if reflect.DeepEqual(original.Icons, copy.Icons) {
		t.Error("Modifying the Icons property of a meta struct must not modify the Icons property of another struct.")
	}
	if reflect.DeepEqual(original.HrefLang, copy.HrefLang) {
		t.Error("Modifying the HrefLang property of a meta struct must not modify the HrefLang property of another struct.")
	}
	if reflect.DeepEqual(original.Alternates, copy.Alternates) {
		t.Error("Modifying the Alternates property of a meta struct must not modify the Alternates property of another struct.")
	}
	if reflect.DeepEqual(original.Stylesheets, copy.Stylesheets) {
		t.Error("Modifying the Stylesheets property of a meta struct must not modify the Stylesheets property of another struct.")
	}
}

func TestJoinKeywords(t *testing.T) {
	original := Meta{
		Keywords: []string{"keyword1", "keyword2", "keyword3"},
	}

	joinedKeywords := original.JoinedKeywords()

	if joinedKeywords != "keyword1, keyword2, keyword3" {
		t.Error("Unable to join the keywords to s string.")
	}
}
