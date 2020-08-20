package authentication

import "github.com/mssola/user_agent"

// Device : struct of a device.
type Device struct {
	Name                 string       `json:"name"`
	Details              string       `json:"details"`
	Type                 string       `json:"type"`
	UDID                 string       `json:"udid,omitempty"`
	RawName              string       `json:"raw_name"`
	NotificationSettings Notification `json:"notification_settings"`
}

// Notification : settings of the notifications on a device.
type Notification struct {
	Token         string   `json:"token,omitempty"`
	Enabled       bool     `json:"enabled"`
	Notifications []string `json:"notifications"`
}

// SetUpNameAndDetails : set ups the name and the details.
func (device *Device) SetUpNameAndDetails() {
	device.RawName = device.Name

	if device.Type == TypeWeb {
		ua := user_agent.New(device.Name)

		name, _ := ua.Browser()
		device.Name = name

		if ua.Bot() {
			device.Details = "bot"
		} else {
			device.Details = ua.OS()
		}
	}
}
