package localization

import "testing"

func TestTranslateFunc(t *testing.T) {
	e := loadEngine(t)
	fr := e.Translate("fr-FR")
	en := e.Translate("en-GB")

	type person struct {
		Name string
	}

	tests := []struct {
		name string
		t    TranslateFunc
		id   string
		args []any
		want string
	}{
		{"simple", fr, "hello", nil, "Bonjour"},
		{"other locale", en, "hello", nil, "Hello"},
		{"map data", fr, "greeting", []any{H{"Name": "Cyril"}}, "Bonjour Cyril"},
		{"string map data", fr, "greeting", []any{map[string]any{"Name": "Cyril"}}, "Bonjour Cyril"},
		{"struct data", fr, "greeting", []any{person{Name: "Cyril"}}, "Bonjour Cyril"},
		{"pointer to struct data", fr, "greeting", []any{&person{Name: "Cyril"}}, "Bonjour Cyril"},
		{"nil pointer data", fr, "hello", []any{(*person)(nil)}, "Bonjour"},
		{"map with non string keys", fr, "hello", []any{map[int]string{1: "x"}}, "Bonjour"},
		{"unsupported data", fr, "hello", []any{42.5, "x", true}, "Bonjour"},
		{"singular count", fr, "items", []any{1}, "1 élément"},
		{"plural count", fr, "items", []any{3}, "3 éléments"},
		{"float count", fr, "items", []any{float64(3)}, "3 éléments"},
		{"float32 count", fr, "items", []any{float32(1)}, "1 élément"},
		{"fractional count", fr, "items", []any{1.5}, "1.5 élément"},
		{"uint count", fr, "items", []any{uint(3)}, "3 éléments"},
		{"uint8 count", fr, "items", []any{uint8(1)}, "1 élément"},
		{"uint16 count", fr, "items", []any{uint16(3)}, "3 éléments"},
		{"uint32 count", fr, "items", []any{uint32(1)}, "1 élément"},
		{"uint64 count", fr, "items", []any{uint64(3)}, "3 éléments"},
		{"string count", fr, "items", []any{"1"}, "1 élément"},
		{"text is the value of Count", fr, "since", []any{"06/11/2026 16:00"}, "Depuis le 06/11/2026 16:00"},
		{"count then data", fr, "greeting", []any{2, H{"Name": "Cyril"}}, "Bonjour Cyril"},
		{"count in data", fr, "items", []any{H{"Count": 1}}, "1 élément"},
		{"missing message returns its ID", fr, "missing", nil, "missing"},
		{"message of another locale is not used", fr, "only_en", nil, "only_en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t(tt.id, tt.args...); got != tt.want {
				t.Errorf("translate(%q, %v) = %q, want %q", tt.id, tt.args, got, tt.want)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"int", 1, true},
		{"int8", int8(1), true},
		{"int16", int16(1), true},
		{"int32", int32(1), true},
		{"int64", int64(1), true},
		{"uint", uint(1), true},
		{"uint8", uint8(1), true},
		{"uint16", uint16(1), true},
		{"uint32", uint32(1), true},
		{"uint64", uint64(1), true},
		{"float32", float32(1), true},
		{"float64", 1.5, true},
		{"numeric string", "12.5", true},
		{"text", "abc", false},
		{"empty string", "", false},
		{"bool", true, false},
		{"nil", nil, false},
		{"map", H{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNumber(tt.value); got != tt.want {
				t.Errorf("isNumber(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestToMap(t *testing.T) {
	type data struct {
		Public  string
		private string
	}

	var nilMap map[string]any
	var iface any = &data{Public: "x"}

	tests := []struct {
		name  string
		value any
		want  map[string]any
	}{
		{"nil", nil, map[string]any{}},
		{"map", map[string]string{"a": "b"}, map[string]any{"a": "b"}},
		{"nil map", nilMap, map[string]any{}},
		{"struct keeps exported fields", data{Public: "x", private: "y"}, map[string]any{"Public": "x"}},
		{"pointer to interface", &iface, map[string]any{"Public": "x"}},
		{"slice", []string{"a"}, map[string]any{}},
		{"non string keys", map[int]int{1: 2}, map[string]any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toMap(tt.value)

			if len(got) != len(tt.want) {
				t.Fatalf("toMap() = %v, want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("toMap()[%q] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}
