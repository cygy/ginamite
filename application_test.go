package ginamite

import (
	"slices"
	"testing"

	"github.com/cygy/ginamite/common/notifications"
)

func TestApplicationConfiguration_Notifications(t *testing.T) {
	a := ApplicationConfiguration{}

	a.AddSupportedNotification("ginamite_test_news", true)
	if !notifications.IsNotificationSupported("ginamite_test_news") || !slices.Contains(notifications.DefaultNotifications(), "ginamite_test_news") {
		t.Fatal("the notification is not supported by default")
	}

	if !a.SetNotificationDefault("ginamite_test_news", false) || slices.Contains(notifications.DefaultNotifications(), "ginamite_test_news") {
		t.Error("the notification is still a default one")
	}
	if !a.DisableNotification("ginamite_test_news") || notifications.IsNotificationSupported("ginamite_test_news") {
		t.Error("the notification is still enabled")
	}
	if !a.EnableNotification("ginamite_test_news") || !notifications.IsNotificationSupported("ginamite_test_news") {
		t.Error("the notification is still disabled")
	}

	if a.DisableNotification("ginamite_test_unknown") || a.EnableNotification("ginamite_test_unknown") || a.SetNotificationDefault("ginamite_test_unknown", true) {
		t.Error("an unknown notification must be reported")
	}
}
