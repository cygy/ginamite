package model

// AdminDevice : struct
type AdminDevice struct {
	Name    string `json:"name" bson:"name"`
	Details string `json:"details" bson:"details"`
	Type    string `json:"type" bson:"type"`
}

// NewAdminDevice : returns a new 'AdminDevice' struct.
func NewAdminDevice(device Device) AdminDevice {
	return AdminDevice{
		Name:    device.Name,
		Details: device.Details,
		Type:    device.Type,
	}
}
