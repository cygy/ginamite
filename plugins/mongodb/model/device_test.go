package model

import (
	"reflect"
	"testing"

	"github.com/cygy/ginamite/common/authentication"
)

func TestDevice_Conversion(t *testing.T) {
	device := Device{
		Name: "phone", Details: "iOS", Type: "mobile", UDID: "u", RawName: "raw",
		NotificationSettings: DeviceNotification{Token: "t", Enabled: true, Notifications: []string{"a", "b"}},
	}

	auth := device.ToDevice()
	want := &authentication.Device{
		Name: "phone", Details: "iOS", Type: "mobile", UDID: "u", RawName: "raw",
		NotificationSettings: authentication.Notification{Token: "t", Enabled: true, Notifications: []string{"a", "b"}},
	}
	if !reflect.DeepEqual(auth, want) {
		t.Errorf("ToDevice() = %+v", auth)
	}

	var back Device
	back.FromDevice(*auth)
	if !reflect.DeepEqual(back, device) {
		t.Errorf("FromDevice() = %+v, want %+v", back, device)
	}
}
