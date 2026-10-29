package model

import (
	"context"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// IPLocation : IP location properties.
type IPLocation struct {
	ID      bson.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	IP      string            `json:"IP" bson:"IP"`
	Region  string            `json:"region" bson:"region"`
	City    string            `json:"city" bson:"city"`
	Country IPLocationCountry `json:"country" bson:"country"`
}

// IPLocationCountry : IP location country properties.
type IPLocationCountry struct {
	Code string `json:"code" bson:"code"`
	Name string `json:"name" bson:"name"`
}

// NewIPLocation : returns a new 'IPLocation' struct.
func NewIPLocation() *IPLocation {
	return &IPLocation{}
}

// IsSaved : returns true if the 'IPLocation' document if saved in the collection.
func (IPLocation *IPLocation) IsSaved() bool {
	return !IPLocation.ID.IsZero()
}

// Save : inserts the 'IPLocation' document in the collection, returns an error if needed.
func (IPLocation *IPLocation) Save(ctx context.Context, db *mongo.Database) (string, error) {
	return document.SaveDocument(ctx, IPLocation, IPLocationCollection, db)
}

// Update : updates the 'IPLocation' document, returns an error if needed.
func (IPLocation *IPLocation) Update(ctx context.Context, update any, db *mongo.Database) error {
	return document.UpdateDocument(ctx, IPLocation.ID, IPLocationCollection, update, db)
}

// Delete : deletes the 'IPLocation' document from the collection, returns an error if needed.
func (IPLocation *IPLocation) Delete(ctx context.Context, db *mongo.Database) error {
	return document.DeleteDocument(ctx, IPLocation.ID, IPLocationCollection, db)
}

// GetByID : initializes the 'IPLocation' struct from the 'IPLocation' document with the ID.
func (IPLocation *IPLocation) GetByID(ctx context.Context, id string, db *mongo.Database) error {
	return document.GetDocumentByID(ctx, IPLocation, IPLocationCollection, id, db)
}

// GetID : returns the 'IPLocation' ID.
func (IPLocation *IPLocation) GetID() string {
	return IPLocation.ID.Hex()
}

// SetID : initializes the 'IPLocation' ID.
func (IPLocation *IPLocation) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.WrongObjectID()
	}
	IPLocation.ID = objectID
	return nil
}
