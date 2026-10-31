package ginamite

import (
	"slices"
	"testing"
)

func TestApplicationConfiguration_Notifications(t *testing.T) {
	s := NewServer()
	a := s.ApplicationConfiguration
	registry := a.Notifications

	a.AddSupportedNotification("ginamite_test_news", true)
	if !registry.IsSupported("ginamite_test_news") || !slices.Contains(registry.Defaults(), "ginamite_test_news") {
		t.Fatal("the notification is not supported by default")
	}

	if !a.SetNotificationDefault("ginamite_test_news", false) || slices.Contains(registry.Defaults(), "ginamite_test_news") {
		t.Error("the notification is still a default one")
	}
	if !a.DisableNotification("ginamite_test_news") || registry.IsSupported("ginamite_test_news") {
		t.Error("the notification is still enabled")
	}
	if !a.EnableNotification("ginamite_test_news") || !registry.IsSupported("ginamite_test_news") {
		t.Error("the notification is still disabled")
	}

	if a.DisableNotification("ginamite_test_unknown") || a.EnableNotification("ginamite_test_unknown") || a.SetNotificationDefault("ginamite_test_unknown", true) {
		t.Error("an unknown notification must be reported")
	}
}

func TestNewServer_SharesTheNotifications(t *testing.T) {
	s := NewServer()

	if s.ApplicationConfiguration.Notifications == nil || s.API.Notifications != s.ApplicationConfiguration.Notifications {
		t.Error("the registry of the notifications must be shared with the API")
	}
}
