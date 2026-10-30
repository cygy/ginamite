package functions

import (
	"context"
	"errors"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/worker/tasks/async/location"

	"go.mongodb.org/mongo-driver/v2/mongo"
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
	if err := IPLocation.GetByIPAddress(ctx, IPAddress, db); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return
	}
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

		if _, err := IPLocation.Save(ctx, db); err != nil {
			log.WithField("error", err).Error("unable to save")
		}
	}

	if err := token.UpdateLocation(ctx, *IPLocation, db); err != nil {
		log.WithField("error", err).Error("unable to update location")
	}
}
