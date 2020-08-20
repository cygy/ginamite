package location

// IPAddressDetails : the details of an IP address.
type IPAddressDetails struct {
	Region      string `json:"region_name"`
	City        string `json:"city"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
}
