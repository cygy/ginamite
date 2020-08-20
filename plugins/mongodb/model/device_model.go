package model

import "github.com/cygy/ginamite/common/authentication"

// Device : struct of a device.
type Device struct {
	Name                 string             `json:"name" bson:"name"`
	Details              string             `json:"details" bson:"details"`
	Type                 string             `json:"type" bson:"type"`
	UDID                 string             `json:"udid,omitempty" bson:"udid,omitempty"`
	RawName              string             `json:"raw_name" bson:"rawName"`
	NotificationSettings DeviceNotification `json:"notification_settings" bson:"notificationSettings"`
}

// DeviceNotification : settings of the notifications on a device.
type DeviceNotification struct {
	Token         string   `json:"token,omitempty" bson:"token,omitempty"`
	Enabled       bool     `json:"enabled" bson:"enabled"`
	Notifications []string `json:"notifications" bson:"notifications"`
}

// ToDevice : returns a struct authentication.Device from this device struct.
func (device *Device) ToDevice() *authentication.Device {
	return &authentication.Device{
		Name:    device.Name,
		Details: device.Details,
		Type:    device.Type,
		UDID:    device.UDID,
		RawName: device.RawName,
		NotificationSettings: authentication.Notification{
			Token:         device.NotificationSettings.Token,
			Enabled:       device.NotificationSettings.Enabled,
			Notifications: device.NotificationSettings.Notifications,
		},
	}
}

// FromDevice : fills this struct from a struct authentication.Device.
func (device *Device) FromDevice(d authentication.Device) {
	device.Name = d.Name
	device.Details = d.Details
	device.Type = d.Type
	device.UDID = d.UDID
	device.RawName = d.RawName
	device.NotificationSettings = DeviceNotification{
		Token:         d.NotificationSettings.Token,
		Enabled:       d.NotificationSettings.Enabled,
		Notifications: d.NotificationSettings.Notifications,
	}
}
