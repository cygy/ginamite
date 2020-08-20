package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

// SaveIPAddressDetails : save the IP address details.
func SaveIPAddressDetails(IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *location.IPAddressDetails) {
	session, db := session.Copy()
	defer session.Close()

	// Verify that the token is still existing.
	token := &model.AuthToken{}
	if err := token.GetByID(tokenID, db); err != nil {
		return
	}

	// Verify the location of the IP address is in cache.
	IPLocation := &model.IPLocation{}
	IPLocation.GetByIPAddress(IPAddress, db)
	if !IPLocation.IsSaved() {
		details := getIPAddressDetailsFunc(IPAddress)
		if details == nil {
			return
		}

		// Save into the db.
		IPLocation.IP = IPAddress
		IPLocation.Region = details.Region
		IPLocation.City = details.City
		IPLocation.Country.Code = details.CountryCode
		IPLocation.Country.Name = details.CountryName

		IPLocation.Save(db)
	}

	token.UpdateLocation(*IPLocation, db)
}
