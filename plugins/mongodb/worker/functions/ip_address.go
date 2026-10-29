package functions

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

// SaveIPAddressDetails : save the IP address details.
func SaveIPAddressDetails(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *location.IPAddressDetails) {
	db := session.Database()

	// Verify that the token is still existing.
	token := &model.AuthToken{}
	if err := token.GetByID(ctx, tokenID, db); err != nil {
		return
	}

	// Verify the location of the IP address is in cache.
	IPLocation := &model.IPLocation{}
	IPLocation.GetByIPAddress(ctx, IPAddress, db)
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

		IPLocation.Save(ctx, db)
	}

	token.UpdateLocation(ctx, *IPLocation, db)
}
