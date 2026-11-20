package model

import (
	"context"
	"strings"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// IPLocationCollection is the name of the collection of the ipLocation documents.
const IPLocationCollection = "IPLocation"

// IPLocationTTL is the number of seconds a location is kept in cache after it is given by the geoip API: it is asked
// again after, and the document is deleted by the TTL index of updatedAt.
const IPLocationTTL = 24 * 3600

// IPLocation is the geographic location of an IP address. RegionCode is ISO 3166-2, such as FR-NAQ. Source is the
// name of the source of the location given by the geoip API, such as icloud-private-relay or dbip; it is empty for
// the locations saved before the API gave it, all from DB-IP. UpdatedAt is the time the API gave the location; it is
// zero for the locations saved before it was kept.
type IPLocation struct {
	ID         bson.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	IP         string            `json:"IP" bson:"IP"`
	Region     string            `json:"region" bson:"region"`
	RegionCode string            `json:"regionCode" bson:"regionCode,omitempty"`
	City       string            `json:"city" bson:"city"`
	Country    IPLocationCountry `json:"country" bson:"country"`
	Source     string            `json:"source" bson:"source,omitempty"`
	UpdatedAt  time.Time         `json:"-" bson:"updatedAt,omitempty"`
}

// IPLocationCountry is the country of an IPLocation.
type IPLocationCountry struct {
	Code string `json:"code" bson:"code"`
	Name string `json:"name" bson:"name"`
}

// Flag returns the emoji of the flag of the country, such as 🇫🇷 for FR, from its ISO 3166-1 alpha-2 code, for the
// templates ({{.Location.Country.Flag}}). It returns an empty string if the code is not two letters.
func (country IPLocationCountry) Flag() string {
	code := strings.ToUpper(country.Code)
	if len(code) != 2 {
		return ""
	}

	// A flag is the pair of the regional indicator symbols of the letters, from U+1F1E6 for A.
	const regionalIndicatorA = 0x1F1E6
	flag := make([]rune, 0, 2)
	for _, letter := range code {
		if letter < 'A' || letter > 'Z' {
			return ""
		}
		flag = append(flag, regionalIndicatorA+letter-'A')
	}

	return string(flag)
}

// NewIPLocation returns a new IPLocation.
func NewIPLocation() *IPLocation {
	return &IPLocation{}
}

// IsFresh reports whether the location was given by the geoip API less than IPLocationTTL seconds before now.
func (location *IPLocation) IsFresh(now time.Time) bool {
	return !location.UpdatedAt.IsZero() && now.Sub(location.UpdatedAt) < IPLocationTTL*time.Second
}

// IsSaved reports whether the IPLocation document is saved in the collection.
func (location *IPLocation) IsSaved() bool {
	return bson.IsObjectIDProvided(location.ID)
}

// Save inserts the IPLocation document in the collection, returning an error if needed.
func (location *IPLocation) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, location, IPLocationCollection, db)
}

// Update updates the IPLocation document, returning an error if needed.
func (location *IPLocation) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, location.ID, IPLocationCollection, update, db)
}

// Delete deletes the IPLocation document from the collection, returning an error if needed.
func (location *IPLocation) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, location.ID, IPLocationCollection, db)
}

// GetByID loads the IPLocation document with the ID.
func (location *IPLocation) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, location, IPLocationCollection, id, db)
}

// GetID returns the IPLocation ID.
func (location *IPLocation) GetID() string {
	return location.ID.Hex()
}

// SetID sets the ID of the IPLocation from its hexadecimal form. The ID is kept if the form is not valid.
func (location *IPLocation) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	location.ID = objectID
	return nil
}

// GetByIPAddress loads the IPLocation document with the IP Address.
func (location *IPLocation) GetByIPAddress(ctx context.Context, IPAddress string, db store.Database) error {
	return document.GetDocumentBySelector(ctx, location, IPLocationCollection, bson.M{"IP": IPAddress}, db)
}

// IsEmpty reports whether this location is empty.
func (location *IPLocation) IsEmpty() bool {
	return len(location.Region) == 0 && len(location.Country.Name) == 0 && len(location.Country.Code) == 0
}
