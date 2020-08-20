package html

import "testing"

func TestPageTemplateCopy(t *testing.T) {
	original := PageTemplate{
		Decorator: "original_decorator",
		Main:      "original_main",
	}

	copy := original

	if original.Decorator != copy.Decorator {
		t.Error("Copying a page template struct must not modify the Decorator property.")
	}
	if original.Main != copy.Main {
		t.Error("Copying a page template struct must not modify the Main property.")
	}

	original.Decorator = original.Decorator + "_modified"
	original.Main = original.Main + "_modified"

	if original.Decorator == copy.Decorator {
		t.Error("Modifying the Decorator property of a page template struct must not modify the Decorator property of the copied struct.")
	}
	if original.Main == copy.Main {
		t.Error("Modifying the Main property of a page template struct must not modify the Main property of the copied struct.")
	}
}

func TestPageTemplateMerge(t *testing.T) {
	original := PageTemplate{
		Decorator: "original_decorator",
		Main:      "original_main",
	}

	copy := PageTemplate{
		Decorator: "copy_decorator",
		Main:      "copy_main",
	}

	original.Merge(copy)

	if original.Decorator != copy.Decorator {
		t.Error("Merging a page template struct must modify the Decorator property.")
	}
	if original.Main != copy.Main {
		t.Error("Merging a page template struct must modify the Main property.")
	}

	original.Decorator = original.Decorator + "_modified"
	original.Main = original.Main + "_modified"

	if original.Decorator == copy.Decorator {
		t.Error("Modifying the Decorator property of a page template struct must not modify the Decorator property of another struct.")
	}
	if original.Main == copy.Main {
		t.Error("Modifying the Main property of a page template struct must not modify the Main property of another struct.")
	}

	copy.Decorator = copy.Decorator + "_copied_and_modified"
	copy.Main = copy.Main + "_copied_and_modified"

	if original.Decorator == copy.Decorator {
		t.Error("Modifying the Decorator property of a page template struct must not modify the Decorator property of another struct.")
	}
	if original.Main == copy.Main {
		t.Error("Modifying the Main property of a page template struct must not modify the Main property of another struct.")
	}
}
