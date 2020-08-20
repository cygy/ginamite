package model

// OwnIPLocation : public version of the IPLocation struct
type OwnIPLocation struct {
	IP      string            `json:"IP"`
	Region  string            `json:"region"`
	City    string            `json:"city"`
	Country IPLocationCountry `json:"country"`
}

// NewOwnIPLocation : returns a new 'OwnIPLocation' struct.
func NewOwnIPLocation(location IPLocation) *OwnIPLocation {
	return &OwnIPLocation{
		IP:      location.IP,
		Region:  location.Region,
		City:    location.City,
		Country: location.Country,
	}
}
