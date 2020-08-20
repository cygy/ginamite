package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// IPLocation : IP location properties.
type IPLocation struct {
	ID      bson.ObjectId     `json:"id,omitempty" bson:"_id,omitempty"`
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
	return IPLocation.ID.Valid()
}

// Save : inserts the 'IPLocation' document in the collection, returns an error if needed.
func (IPLocation *IPLocation) Save(db *mgo.Database) (string, error) {
	return document.SaveDocument(IPLocation, IPLocationCollection, db)
}

// Update : updates the 'IPLocation' document, returns an error if needed.
func (IPLocation *IPLocation) Update(update interface{}, db *mgo.Database) error {
	return document.UpdateDocument(IPLocation.ID, IPLocationCollection, update, db)
}

// Delete : deletes the 'IPLocation' document from the collection, returns an error if needed.
func (IPLocation *IPLocation) Delete(db *mgo.Database) error {
	return document.DeleteDocument(IPLocation.ID, IPLocationCollection, db)
}

// GetByID : initializes the 'IPLocation' struct from the 'IPLocation' document with the ID.
func (IPLocation *IPLocation) GetByID(id string, db *mgo.Database) error {
	return document.GetDocumentByID(IPLocation, IPLocationCollection, id, db)
}

// GetID : returns the 'IPLocation' ID.
func (IPLocation *IPLocation) GetID() string {
	return IPLocation.ID.Hex()
}

// SetID : initializes the 'IPLocation' ID.
func (IPLocation *IPLocation) SetID(id string) {
	IPLocation.ID = bson.ObjectIdHex(id)
}
