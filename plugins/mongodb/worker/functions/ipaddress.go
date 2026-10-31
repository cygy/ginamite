package functions

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

// SaveIPAddressDetails saves the location of the IP address of an auth token, and updates the location of the token.
// The location is read from the database if it is known, and from getIPAddressDetailsFunc otherwise. Nothing is
// done if the token does not exist anymore or if the location is not found. It returns the error of the database, if
// any: it can be called again.
func (f *Functions) SaveIPAddressDetails(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *location.IPAddressDetails) error {
	db := f.database()

	// Verify that the token is still existing.
	token := &model.AuthToken{}
	if err := token.GetByID(ctx, tokenID, db); err != nil {
		if documentMissing(err) {
			return nil
		}

		return fmt.Errorf("unable to get the auth token: %w", err)
	}

	// Verify the location of the IP address is in cache.
	location := &model.IPLocation{}
	if err := location.GetByIPAddress(ctx, IPAddress, db); err != nil && !documentMissing(err) {
		return fmt.Errorf("unable to get the location of the IP address: %w", err)
	}
	if !location.IsSaved() {
		details := getIPAddressDetailsFunc(IPAddress)
		if details == nil {
			return nil
		}

		// Save into the db.
		location.IP = IPAddress
		location.Region = details.Region
		location.City = details.City
		location.Country.Code = details.CountryCode
		location.Country.Name = details.CountryName

		if _, err := location.Save(ctx, db); err != nil {
			return fmt.Errorf("unable to save the location of the IP address: %w", err)
		}
	}

	if err := token.UpdateLocation(ctx, *location, db); err != nil && !documentMissing(err) {
		return fmt.Errorf("unable to update the location of the auth token: %w", err)
	}

	return nil
}
