package model

import "testing"

func TestNewAdminDevice(t *testing.T) {
	got := NewAdminDevice(Device{Name: "phone", Details: "iOS", Type: "mobile", UDID: "secret", NotificationSettings: DeviceNotification{Token: "secret"}})
	if got != (AdminDevice{Name: "phone", Details: "iOS", Type: "mobile"}) {
		t.Errorf("got %+v", got)
	}
}
