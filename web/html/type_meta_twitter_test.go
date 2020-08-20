package html

import "testing"

func TestMetaTwitterCopy(t *testing.T) {
	original := Twitter{
		Card:     "original_card",
		ImageAlt: "original_image_alt",
		Site:     "original_site",
		Creator:  "original_creator",
	}

	copy := original

	if original.Card != copy.Card {
		t.Error("Copying a meta twitter struct must not modify the Card property.")
	}
	if original.ImageAlt != copy.ImageAlt {
		t.Error("Copying a meta twitter struct must not modify the ImageAlt property.")
	}
	if original.Site != copy.Site {
		t.Error("Copying a meta twitter struct must not modify the Site property.")
	}
	if original.Creator != copy.Creator {
		t.Error("Copying a meta twitter struct must not modify the Creator property.")
	}

	original.Card = original.Card + "_modified"
	original.ImageAlt = original.ImageAlt + "_modified"
	original.Site = original.Site + "_modified"
	original.Creator = original.Creator + "_modified"

	if original.Card == copy.Card {
		t.Error("Modifying the Card property of a meta twitter struct must not modify the Card property of the copied struct.")
	}
	if original.ImageAlt == copy.ImageAlt {
		t.Error("Modifying the ImageAlt property of a meta twitter struct must not modify the ImageAlt property of the copied struct.")
	}
	if original.Site == copy.Site {
		t.Error("Modifying the Site property of a meta twitter struct must not modify the Site property of the copied struct.")
	}
	if original.Creator == copy.Creator {
		t.Error("Modifying the Creator property of a meta twitter struct must not modify the Creator property of the copied struct.")
	}
}

func TestMetaTwitterMerge(t *testing.T) {
	original := Twitter{
		Card:     "original_card",
		ImageAlt: "original_image_alt",
		Site:     "original_site",
		Creator:  "original_creator",
	}

	copy := Twitter{
		Card:     "copy_card",
		ImageAlt: "copy_image_alt",
		Site:     "copy_site",
		Creator:  "copy_creator",
	}

	original.Merge(copy)

	if original.Card != copy.Card {
		t.Error("Merging a meta twitter struct must modify the Card property.")
	}
	if original.ImageAlt != copy.ImageAlt {
		t.Error("Merging a meta twitter struct must modify the ImageAlt property.")
	}
	if original.Site != copy.Site {
		t.Error("Merging a meta twitter struct must modify the Site property.")
	}
	if original.Creator != copy.Creator {
		t.Error("Merging a meta twitter struct must modify the Creator property.")
	}

	original.Card = original.Card + "_modified"
	original.ImageAlt = original.ImageAlt + "_modified"
	original.Site = original.Site + "_modified"
	original.Creator = original.Creator + "_modified"

	if original.Card == copy.Card {
		t.Error("Modifying the Card property of a meta twitter struct must not modify the Card property of another struct.")
	}
	if original.ImageAlt == copy.ImageAlt {
		t.Error("Modifying the ImageAlt property of a meta twitter struct must not modify the ImageAlt property of another struct.")
	}
	if original.Site == copy.Site {
		t.Error("Modifying the Site property of a meta twitter struct must not modify the Site property of another struct.")
	}
	if original.Creator == copy.Creator {
		t.Error("Modifying the Creator property of a meta twitter struct must not modify the Creator property of the copied struct.")
	}

	copy.Card = copy.Card + "_copied_and_modified"
	copy.ImageAlt = copy.ImageAlt + "_copied_and_modified"
	copy.Site = copy.Site + "_copied_and_modified"
	copy.Creator = copy.Creator + "_copied_and_modified"

	if original.Card == copy.Card {
		t.Error("Modifying the Card property of a meta twitter struct must not modify the Card property of another struct.")
	}
	if original.ImageAlt == copy.ImageAlt {
		t.Error("Modifying the ImageAlt property of a meta twitter struct must not modify the ImageAlt property of another struct.")
	}
	if original.Site == copy.Site {
		t.Error("Modifying the Site property of a meta twitter struct must not modify the Site property of another struct.")
	}
	if original.Creator == copy.Creator {
		t.Error("Modifying the Creator property of a meta twitter struct must not modify the Creator property of the copied struct.")
	}
}
