package html

import "testing"

func TestMetaStylesheetCopy(t *testing.T) {
	original := Stylesheet{
		Href:  "original_href",
		Media: "original_media",
	}

	copy := original

	if original.Href != copy.Href {
		t.Error("Copying a meta stylesheet struct must not modify the Href property.")
	}
	if original.Media != copy.Media {
		t.Error("Copying a meta stylesheet struct must not modify the Media property.")
	}

	original.Href = original.Href + "_modified"
	original.Media = original.Media + "_modified"

	if original.Href == copy.Href {
		t.Error("Modifying the Href property of a meta stylesheet struct must not modify the Href property of the copied struct.")
	}
	if original.Media == copy.Media {
		t.Error("Modifying the Media property of a meta stylesheet struct must not modify the Media property of the copied struct.")
	}
}
