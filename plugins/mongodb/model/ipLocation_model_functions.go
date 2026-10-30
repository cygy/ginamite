package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// GetByIPAddress : initializes the 'IPLocation' struct from the 'IPLocation' document with the IP Address.
func (location *IPLocation) GetByIPAddress(ctx context.Context, IPAddress string, db store.Database) error {
	return document.GetDocumentBySelector(ctx, location, IPLocationCollection, bson.M{"IP": IPAddress}, db)
}

// IsEmpty : returns true if this location is empty.
func (location *IPLocation) IsEmpty() bool {
	return len(location.Region) == 0 && len(location.Country.Name) == 0 && len(location.Country.Code) == 0
}
