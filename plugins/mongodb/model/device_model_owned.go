package model

// OwnDevice : struct
type OwnDevice struct {
	Name                 string                `json:"name" bson:"name"`
	Details              string                `json:"details" bson:"details"`
	Type                 string                `json:"type" bson:"type"`
	NotificationSettings OwnDeviceNotification `json:"notification_settings" bson:"notificationSettings"`
}

// OwnDeviceNotification : struct
type OwnDeviceNotification struct {
	Enabled       bool     `json:"enabled" bson:"enabled"`
	Notifications []string `json:"notifications" bson:"notifications"`
}

// NewOwnDevice : returns a new 'OwnDevice' struct.
func NewOwnDevice(device Device) OwnDevice {
	return OwnDevice{
		Name:                 device.Name,
		Details:              device.Details,
		Type:                 device.Type,
		NotificationSettings: NewOwnDeviceNotification(device.NotificationSettings),
	}
}

// NewOwnDeviceNotification : returns a new 'OwnDeviceNotification' struct.
func NewOwnDeviceNotification(deviceNotification DeviceNotification) OwnDeviceNotification {
	return OwnDeviceNotification{
		Enabled:       deviceNotification.Enabled,
		Notifications: deviceNotification.Notifications,
	}
}
