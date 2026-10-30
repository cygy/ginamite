package image

import "testing"

func TestProperty_DefineNameAndPaths(t *testing.T) {
	var p Property
	if p.IsDefined() {
		t.Error("an empty property must not be defined")
	}

	p.DefineNameAndPaths("photo", ".png", "-small", "@2x", "/dest")

	if want := "photo-small@2x.png"; p.Name != want {
		t.Errorf("Name = %q, want %q", p.Name, want)
	}
	if want := "p/h/photo-small@2x.png"; p.RelativePath != want {
		t.Errorf("RelativePath = %q, want %q", p.RelativePath, want)
	}
	if want := "/dest/p/h/photo-small@2x.png"; p.AbsolutePath != want {
		t.Errorf("AbsolutePath = %q, want %q", p.AbsolutePath, want)
	}
	if !p.IsDefined() {
		t.Error("a property with a name must be defined")
	}
}
