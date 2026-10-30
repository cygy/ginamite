package model

import (
	"reflect"
	"testing"
)

func TestNewOwnDevice(t *testing.T) {
	device := Device{Name: "phone", Details: "iOS", Type: "mobile", UDID: "secret", NotificationSettings: DeviceNotification{Token: "secret", Enabled: true, Notifications: []string{"a"}}}

	got := NewOwnDevice(device)
	want := OwnDevice{Name: "phone", Details: "iOS", Type: "mobile", NotificationSettings: OwnDeviceNotification{Enabled: true, Notifications: []string{"a"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v", got)
	}
}

func TestNewOwnDeviceNotification(t *testing.T) {
	got := NewOwnDeviceNotification(DeviceNotification{Token: "secret", Enabled: true, Notifications: []string{"a"}})
	if !got.Enabled || len(got.Notifications) != 1 {
		t.Errorf("got %+v", got)
	}
}
