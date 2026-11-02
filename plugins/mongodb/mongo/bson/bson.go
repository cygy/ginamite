// Package bson re-exports the BSON types and helpers of the MongoDB driver used by the application.
package bson

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// BSON types of the driver.
type (
	// D is an ordered BSON document.
	D = bson.D

	// E is an element of a D.
	E = bson.E

	// M is an unordered BSON document.
	M = bson.M

	// ObjectID is the BSON ObjectID type.
	ObjectID = bson.ObjectID
)

// ObjectIDFromHex creates an ObjectID from its hexadecimal representation.
func ObjectIDFromHex(s string) (bson.ObjectID, error) {
	return bson.ObjectIDFromHex(s)
}

// NewObjectID generates a new ObjectID.
func NewObjectID() bson.ObjectID {
	return bson.NewObjectID()
}

// IsNilObjectID reports whether the hexadecimal ObjectID is empty or made of zeros, i.e. not provided.
func IsNilObjectID(objectID string) bool {
	return len(objectID) == 0 || objectID == bson.NilObjectID.Hex()
}

// IsObjectIDProvided reports whether the ObjectID is not the zero value.
func IsObjectIDProvided(objectID bson.ObjectID) bool {
	return !objectID.IsZero()
}
