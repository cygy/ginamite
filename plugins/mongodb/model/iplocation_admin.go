package model

// AdminIPLocation is the view of an IPLocation exposed to the administrators.
type AdminIPLocation struct {
	IP         string            `json:"IP"`
	Region     string            `json:"region"`
	RegionCode string            `json:"regionCode"`
	City       string            `json:"city"`
	Country    IPLocationCountry `json:"country"`
	Source     string            `json:"source"`
}

// NewAdminIPLocation returns a new AdminIPLocation.
func NewAdminIPLocation(location IPLocation) *AdminIPLocation {
	return &AdminIPLocation{
		IP:         location.IP,
		Region:     location.Region,
		RegionCode: location.RegionCode,
		City:       location.City,
		Country:    location.Country,
		Source:     location.Source,
	}
}

// IsEmpty reports whether this location is empty.
func (location *AdminIPLocation) IsEmpty() bool {
	return len(location.Region) == 0 && len(location.Country.Name) == 0
}
