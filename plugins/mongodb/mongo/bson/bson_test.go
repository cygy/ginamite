package bson

import "testing"

func TestObjectIDFromHex(t *testing.T) {
	id := NewObjectID()

	got, err := ObjectIDFromHex(id.Hex())
	if err != nil || got != id {
		t.Errorf("ObjectIDFromHex() = %v, %v, want %v", got, err, id)
	}

	if _, err := ObjectIDFromHex("nope"); err == nil {
		t.Error("an invalid hex string must be rejected")
	}
}

func TestNewObjectID(t *testing.T) {
	a, b := NewObjectID(), NewObjectID()
	if a == b || !IsObjectIDProvided(a) {
		t.Errorf("NewObjectID() = %v, %v", a, b)
	}
}

func TestIsNilObjectID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"empty", "", true},
		{"zero object ID", ObjectID{}.Hex(), true},
		{"generated", NewObjectID().Hex(), false},
		{"any text", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNilObjectID(tt.id); got != tt.want {
				t.Errorf("IsNilObjectID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestIsObjectIDProvided(t *testing.T) {
	if IsObjectIDProvided(ObjectID{}) || !IsObjectIDProvided(NewObjectID()) {
		t.Error("IsObjectIDProvided() is wrong")
	}
}

func TestAliases(t *testing.T) {
	d := D{E{Key: "a", Value: 1}}
	m := M{"a": 1}

	if len(d) != 1 || m["a"] != 1 {
		t.Error("the aliases must behave like the driver types")
	}
}
