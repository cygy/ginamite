package model

// AdminIPLocation : admin version of the IPLocation struct
type AdminIPLocation struct {
	IP      string            `json:"IP"`
	Region  string            `json:"region"`
	City    string            `json:"city"`
	Country IPLocationCountry `json:"country"`
}

// NewAdminIPLocation : returns a new 'AdminIPLocation' struct.
func NewAdminIPLocation(location IPLocation) *AdminIPLocation {
	return &AdminIPLocation{
		IP:      location.IP,
		Region:  location.Region,
		City:    location.City,
		Country: location.Country,
	}
}
