package model

// OwnIPLocation is the view of an IPLocation exposed to its owner.
type OwnIPLocation struct {
	IP      string            `json:"IP"`
	Region  string            `json:"region"`
	City    string            `json:"city"`
	Country IPLocationCountry `json:"country"`
}

// NewOwnIPLocation returns a new OwnIPLocation.
func NewOwnIPLocation(location IPLocation) *OwnIPLocation {
	return &OwnIPLocation{
		IP:      location.IP,
		Region:  location.Region,
		City:    location.City,
		Country: location.Country,
	}
}

// IsEmpty reports whether this location is empty.
func (location *OwnIPLocation) IsEmpty() bool {
	return len(location.Region) == 0 && len(location.Country.Name) == 0
}
