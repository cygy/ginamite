package html

import "testing"

func TestMetaHrefLangCopy(t *testing.T) {
	original := HrefLang{
		Href:   "original_href",
		Locale: "original_locale",
	}

	copy := original

	if original.Href != copy.Href {
		t.Error("Copying a meta hreflang struct must not modify the Href property.")
	}
	if original.Locale != copy.Locale {
		t.Error("Copying a meta hreflang struct must not modify the Locale property.")
	}

	original.Href = original.Href + "_modified"
	original.Locale = original.Locale + "_modified"

	if original.Href == copy.Href {
		t.Error("Modifying the Href property of a meta hreflang struct must not modify the Href property of the copied struct.")
	}
	if original.Locale == copy.Locale {
		t.Error("Modifying the Locale property of a meta hreflang struct must not modify the Locale property of the copied struct.")
	}
}
