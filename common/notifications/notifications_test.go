package notifications

import (
	"reflect"
	"strings"
	"testing"
)

// resetNotifications restores the supported notifications after the test.
func resetNotifications(t *testing.T) {
	t.Helper()

	saved := append([]notificationType(nil), supportedNotifications...)
	t.Cleanup(func() { supportedNotifications = saved })
}

func TestIsNotificationSupported(t *testing.T) {
	resetNotifications(t)

	tests := []struct {
		name         string
		notification string
		prepare      func()
		want         bool
	}{
		{"default notification", TypeNewLogin, func() {}, true},
		{"unknown notification", "unknown", func() {}, false},
		{"disabled notification", TypeNewLogin, func() { DisableNotification(TypeNewLogin) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetNotifications(t)
			tt.prepare()

			if got := IsNotificationSupported(tt.notification); got != tt.want {
				t.Errorf("IsNotificationSupported(%q) = %v, want %v", tt.notification, got, tt.want)
			}
		})
	}
}

func TestAddNotification(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		resetNotifications(t)
		AddNotification("news", true)

		if !IsNotificationSupported("news") || !reflect.DeepEqual(DefaultNotifications(), []string{TypeNewLogin, "news"}) {
			t.Errorf("supported = %v, default = %v", SupportedNotifications(), DefaultNotifications())
		}
	})

	t.Run("existing is enabled again and updated", func(t *testing.T) {
		resetNotifications(t)
		DisableNotification(TypeCommercialPartners)
		AddNotification(TypeCommercialPartners, true)

		if !IsNotificationSupported(TypeCommercialPartners) {
			t.Error("the notification must be enabled again")
		}
		if !reflect.DeepEqual(DefaultNotifications(), []string{TypeNewLogin, TypeCommercialPartners}) {
			t.Errorf("default = %v", DefaultNotifications())
		}
		if len(supportedNotifications) != 2 {
			t.Errorf("the notification must not be duplicated: %v", supportedNotifications)
		}
	})
}

func TestEnableAndDisableNotification(t *testing.T) {
	resetNotifications(t)

	if !DisableNotification(TypeNewLogin) || IsNotificationSupported(TypeNewLogin) {
		t.Error("DisableNotification() must disable a known notification")
	}
	if !EnableNotification(TypeNewLogin) || !IsNotificationSupported(TypeNewLogin) {
		t.Error("EnableNotification() must enable a known notification")
	}
	if EnableNotification("unknown") || DisableNotification("unknown") {
		t.Error("an unknown notification must be reported")
	}
}

func TestSetNotificationDefault(t *testing.T) {
	resetNotifications(t)

	if !SetNotificationDefault(TypeNewLogin, false) || len(DefaultNotifications()) != 0 {
		t.Errorf("default = %v", DefaultNotifications())
	}
	if !SetNotificationDefault(TypeCommercialPartners, true) || !reflect.DeepEqual(DefaultNotifications(), []string{TypeCommercialPartners}) {
		t.Errorf("default = %v", DefaultNotifications())
	}
	if SetNotificationDefault("unknown", true) {
		t.Error("an unknown notification must be reported")
	}
}

func TestSupportedAndDefaultNotifications(t *testing.T) {
	resetNotifications(t)

	if got := SupportedNotifications(); !reflect.DeepEqual(got, []string{TypeNewLogin, TypeCommercialPartners}) {
		t.Errorf("SupportedNotifications() = %v", got)
	}
	if got := DefaultNotifications(); !reflect.DeepEqual(got, []string{TypeNewLogin}) {
		t.Errorf("DefaultNotifications() = %v", got)
	}

	DisableNotification(TypeNewLogin)
	if got := SupportedNotifications(); !reflect.DeepEqual(got, []string{TypeCommercialPartners}) {
		t.Errorf("SupportedNotifications() = %v", got)
	}
	if got := DefaultNotifications(); got == nil || len(got) != 0 {
		t.Errorf("DefaultNotifications() = %#v, want an empty slice", got)
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
