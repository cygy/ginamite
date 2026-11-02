package model

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// IPLocationCollection is the name of the collection of the ipLocation documents.
const IPLocationCollection = "IPLocation"

// IPLocation is the geographic location of an IP address.
type IPLocation struct {
	ID      bson.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	IP      string            `json:"IP" bson:"IP"`
	Region  string            `json:"region" bson:"region"`
	City    string            `json:"city" bson:"city"`
	Country IPLocationCountry `json:"country" bson:"country"`
}

// IPLocationCountry is the country of an IPLocation.
type IPLocationCountry struct {
	Code string `json:"code" bson:"code"`
	Name string `json:"name" bson:"name"`
}

// NewIPLocation returns a new IPLocation.
func NewIPLocation() *IPLocation {
	return &IPLocation{}
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
