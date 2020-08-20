package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetByIPAddress : initializes the 'IPLocation' struct from the 'IPLocation' document with the IP Address.
func (location *IPLocation) GetByIPAddress(IPAddress string, db *mgo.Database) error {
	return document.GetDocumentBySelector(location, IPLocationCollection, bson.M{"IP": IPAddress}, db)
}

// IsEmpty : returns true if this location is empty.
func (location *IPLocation) IsEmpty() bool {
	return len(location.Region) == 0 && len(location.Country.Name) == 0 && len(location.Country.Code) == 0
}
