package authentication

import "github.com/mssola/useragent"

// Device describes the device used to authenticate.
type Device struct {
	Name                 string       `json:"name"`
	Details              string       `json:"details"`
	Type                 string       `json:"type"`
	UDID                 string       `json:"udid,omitempty"`
	RawName              string       `json:"raw_name"`
	NotificationSettings Notification `json:"notification_settings"`
}

// Notification holds the settings of the notifications on a device.
type Notification struct {
	Token         string   `json:"token,omitempty"`
	Enabled       bool     `json:"enabled"`
	Notifications []string `json:"notifications"`
}

// SetUpNameAndDetails keeps the raw name and, for a web device, derives the
// browser name and the details (operating system, or "bot") from the user agent.
func (device *Device) SetUpNameAndDetails() {
	device.RawName = device.Name

	if device.Type == TypeWeb {
		ua := useragent.New(device.Name)

		name, _ := ua.Browser()
		device.Name = name

		if ua.Bot() {
			device.Details = "bot"
		} else {
			device.Details = ua.OS()
		}
	}
}
