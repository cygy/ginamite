package functions

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

// SaveIPAddressDetails saves the location of the IP address of an auth token, and updates the location of the token.
// The location is read from the database if it was given by getIPAddressDetailsFunc less than model.IPLocationTTL
// seconds before, and from getIPAddressDetailsFunc otherwise: an older location is asked again and updated, and kept
// if the API finds nothing. Nothing is done if the token does not exist anymore or if the location is unknown. It returns the error of
// the database or of getIPAddressDetailsFunc, if any: it can be called again. The function of the geoip API returns
// none: it only logs its failures.
func (f *Functions) SaveIPAddressDetails(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc location.DetailsFunc) error {
	db := f.database()

	// Verify that the token is still existing.
	token := &model.AuthToken{}
	if err := token.GetByID(ctx, tokenID, db); err != nil {
		if documentMissing(err) {
			return nil
		}

		return fmt.Errorf("unable to get the auth token: %w", err)
	}

	// Verify the location of the IP address is in cache, and fresh.
	location := &model.IPLocation{}
	if err := location.GetByIPAddress(ctx, IPAddress, db); err != nil && !documentMissing(err) {
		return fmt.Errorf("unable to get the location of the IP address: %w", err)
	}
	if now := time.Now(); !location.IsSaved() || !location.IsFresh(now) {
		details, err := getIPAddressDetailsFunc(IPAddress)
		if err != nil {
			return fmt.Errorf("unable to get the location of the IP address: %w", err)
		}
		if details == nil && !location.IsSaved() {
			return nil
		}
		if details != nil {
			if err := saveLocation(ctx, location, IPAddress, details, now, db); err != nil {
				return err
			}
		}
	}

	if err := token.UpdateLocation(ctx, *location, db); err != nil && !documentMissing(err) {
		return fmt.Errorf("unable to update the location of the auth token: %w", err)
	}

	return nil
}

// saveLocation sets the details of the IP address, given at now, in the location, then inserts it, or updates it if it
// is saved.
func saveLocation(ctx context.Context, ipLocation *model.IPLocation, IPAddress string, details *location.IPAddressDetails, now time.Time, db store.Database) error {
	ipLocation.IP = IPAddress
	ipLocation.Region = details.Region
	ipLocation.RegionCode = details.RegionCode
	ipLocation.City = details.City
	ipLocation.Country.Code = details.CountryCode
	ipLocation.Country.Name = details.CountryName
	ipLocation.Source = details.Source
	ipLocation.UpdatedAt = now

	if ipLocation.IsSaved() {
		update := bson.M{"$set": bson.M{
			"region": ipLocation.Region, "regionCode": ipLocation.RegionCode, "city": ipLocation.City,
			"country": ipLocation.Country, "source": ipLocation.Source, "updatedAt": ipLocation.UpdatedAt,
		}}
		if err := ipLocation.Update(ctx, update, db); err != nil && !documentMissing(err) {
			return fmt.Errorf("unable to update the location of the IP address: %w", err)
		}

		return nil
	}

	if _, err := ipLocation.Save(ctx, db); err != nil {
		return fmt.Errorf("unable to save the location of the IP address: %w", err)
	}

	return nil
}
