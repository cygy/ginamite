package html

import "testing"

func TestMetaIconCopy(t *testing.T) {
	original := Icon{
		Rel:    "original_rel",
		Href:   "original_href",
		Type:   "original_type",
		Color:  "original_color",
		Width:  1,
		Height: 1,
	}

	copy := original

	if original.Rel != copy.Rel {
		t.Error("Copying a meta Icon struct must not modify the Rel property.")
	}
	if original.Href != copy.Href {
		t.Error("Copying a meta Icon struct must not modify the Href property.")
	}
	if original.Type != copy.Type {
		t.Error("Copying a meta Icon struct must not modify the Type property.")
	}
	if original.Color != copy.Color {
		t.Error("Copying a meta Icon struct must not modify the Color property.")
	}
	if original.Width != copy.Width {
		t.Error("Copying a meta Icon struct must not modify the Width property.")
	}
	if original.Height != copy.Height {
		t.Error("Copying a meta Icon struct must not modify the Height property.")
	}

	original.Rel = original.Rel + "_modified"
	original.Href = original.Href + "_modified"
	original.Type = original.Type + "_modified"
	original.Color = original.Color + "_modified"
	original.Width++
	original.Height++

	if original.Rel == copy.Rel {
		t.Error("Modifying the Rel property of a meta Icon struct must not modify the Rel property of the copied struct.")
	}
	if original.Href == copy.Href {
		t.Error("Modifying the Href property of a meta Icon struct must not modify the Href property of the copied struct.")
	}
	if original.Type == copy.Type {
		t.Error("Modifying the Type property of a meta Icon struct must not modify the Type property of the copied struct.")
	}
	if original.Color == copy.Color {
		t.Error("Modifying the Color property of a meta Icon struct must not modify the Color property of the copied struct.")
	}
	if original.Width == copy.Width {
		t.Error("Modifying the Width property of a meta Icon struct must not modify the Width property of the copied struct.")
	}
	if original.Height == copy.Height {
		t.Error("Modifying the Height property of a meta Icon struct must not modify the Height property of the copied struct.")
	}
}
