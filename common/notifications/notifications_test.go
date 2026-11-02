package notifications

import (
	"reflect"
	"strings"
	"testing"
)

func TestRegistry_IsSupported(t *testing.T) {
	tests := []struct {
		name         string
		notification string
		prepare      func(r *Registry)
		want         bool
	}{
		{"default notification", TypeNewLogin, func(*Registry) {}, true},
		{"unknown notification", "unknown", func(*Registry) {}, false},
		{"disabled notification", TypeNewLogin, func(r *Registry) { r.Disable(TypeNewLogin) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.prepare(r)

			if got := r.IsSupported(tt.notification); got != tt.want {
				t.Errorf("r.IsSupported(%q) = %v, want %v", tt.notification, got, tt.want)
			}
		})
	}
}

func TestRegistry_Add(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		r := NewRegistry()
		r.Add("news", true)

		if !r.IsSupported("news") || !reflect.DeepEqual(r.Defaults(), []string{TypeNewLogin, "news"}) {
			t.Errorf("supported = %v, default = %v", r.Supported(), r.Defaults())
		}
	})

	t.Run("existing is enabled again and updated", func(t *testing.T) {
		r := NewRegistry()
		r.Disable(TypeCommercialPartners)
		r.Add(TypeCommercialPartners, true)

		if !r.IsSupported(TypeCommercialPartners) {
			t.Error("the notification must be enabled again")
		}
		if !reflect.DeepEqual(r.Defaults(), []string{TypeNewLogin, TypeCommercialPartners}) {
			t.Errorf("default = %v", r.Defaults())
		}
		if len(r.supported) != 2 {
			t.Errorf("the notification must not be duplicated: %v", r.supported)
		}
	})
}

func TestRegistry_EnableAndDisable(t *testing.T) {
	r := NewRegistry()

	if !r.Disable(TypeNewLogin) || r.IsSupported(TypeNewLogin) {
		t.Error("Disable() must disable a known notification")
	}
	if !r.Enable(TypeNewLogin) || !r.IsSupported(TypeNewLogin) {
		t.Error("Enable() must enable a known notification")
	}
	if r.Enable("unknown") || r.Disable("unknown") {
		t.Error("an unknown notification must be reported")
	}
}

func TestRegistry_SetDefault(t *testing.T) {
	r := NewRegistry()

	if !r.SetDefault(TypeNewLogin, false) || len(r.Defaults()) != 0 {
		t.Errorf("default = %v", r.Defaults())
	}
	if !r.SetDefault(TypeCommercialPartners, true) || !reflect.DeepEqual(r.Defaults(), []string{TypeCommercialPartners}) {
		t.Errorf("default = %v", r.Defaults())
	}
	if r.SetDefault("unknown", true) {
		t.Error("an unknown notification must be reported")
	}
}

func TestRegistry_SupportedAndDefaults(t *testing.T) {
	r := NewRegistry()

	if got := r.Supported(); !reflect.DeepEqual(got, []string{TypeNewLogin, TypeCommercialPartners}) {
		t.Errorf("Supported() = %v", got)
	}
	if got := r.Defaults(); !reflect.DeepEqual(got, []string{TypeNewLogin}) {
		t.Errorf("Defaults() = %v", got)
	}

	r.Disable(TypeNewLogin)
	if got := r.Supported(); !reflect.DeepEqual(got, []string{TypeCommercialPartners}) {
		t.Errorf("Supported() = %v", got)
	}
	if got := r.Defaults(); got == nil || len(got) != 0 {
		t.Errorf("Defaults() = %#v, want an empty slice", got)
	}
}

func TestGenerateUnsubscribeKey(t *testing.T) {
	key := GenerateUnsubscribeKey("user", "secret", TypeNewLogin)

	parts := strings.Split(key, ".")
	if len(parts) != 3 || parts[0] != "user" || parts[1] != TypeNewLogin || len(parts[2]) != 64 {
		t.Fatalf("key = %q", key)
	}
	if key != GenerateUnsubscribeKey("user", "secret", TypeNewLogin) {
		t.Error("the key must be deterministic")
	}
	if key == GenerateUnsubscribeKey("user", "other", TypeNewLogin) || key == GenerateUnsubscribeKey("user", "secret", TypeCommercialPartners) {
		t.Error("the key must depend on the private key and the notification")
	}
}
