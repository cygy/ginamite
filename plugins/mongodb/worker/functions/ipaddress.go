package functions

import (
	"context"
	"errors"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/worker/tasks/async/location"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SaveIPAddressDetails saves the location of the IP address of an auth token, and updates the location of the token.
// The location is read from the database if it is known, and from getIPAddressDetailsFunc otherwise. Nothing is
// done if the token does not exist anymore or if the location is not found.
func (f *Functions) SaveIPAddressDetails(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *location.IPAddressDetails) {
	db := f.database()

	// Verify that the token is still existing.
	token := &model.AuthToken{}
	if err := token.GetByID(ctx, tokenID, db); err != nil {
		return
	}

	// Verify the location of the IP address is in cache.
	location := &model.IPLocation{}
	if err := location.GetByIPAddress(ctx, IPAddress, db); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return
	}
	if !location.IsSaved() {
		details := getIPAddressDetailsFunc(IPAddress)
		if details == nil {
			return
		}

		// Save into the db.
		location.IP = IPAddress
		location.Region = details.Region
		location.City = details.City
		location.Country.Code = details.CountryCode
		location.Country.Name = details.CountryName

		if _, err := location.Save(ctx, db); err != nil {
			log.WithField("error", err).Error("unable to save")
		}
	}

	if err := token.UpdateLocation(ctx, *location, db); err != nil {
		log.WithField("error", err).Error("unable to update location")
	}
}
