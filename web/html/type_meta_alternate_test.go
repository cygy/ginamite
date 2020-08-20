package html

import "testing"

func TestMetaAlternateCopy(t *testing.T) {
	original := Alternate{
		Href:  "original_href",
		Type:  "original_type",
		Title: "original_title",
	}

	copy := original

	if original.Href != copy.Href {
		t.Error("Copying a meta alternate struct must not modify the Href property.")
	}
	if original.Type != copy.Type {
		t.Error("Copying a meta alternate struct must not modify the Type property.")
	}
	if original.Title != copy.Title {
		t.Error("Copying a meta alternate struct must not modify the Title property.")
	}

	original.Href = original.Href + "_modified"
	original.Type = original.Href + "_modified"
	original.Title = original.Href + "_modified"

	if original.Href == copy.Href {
		t.Error("Modifying the Href property of a meta alternate struct must not modify the Href property of the copied struct.")
	}
	if original.Type == copy.Type {
		t.Error("Modifying the Type property of a meta alternate struct must not modify the Type property of the copied struct.")
	}
	if original.Title == copy.Title {
		t.Error("Modifying the Title property of a meta alternate struct must not modify the Title property of the copied struct.")
	}
}
