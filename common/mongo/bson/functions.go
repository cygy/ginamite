package bson

import "go.mongodb.org/mongo-driver/v2/bson"

func ObjectIDFromHex(s string) (bson.ObjectID, error) {
	return bson.ObjectIDFromHex(s)
}

func NewObjectID() bson.ObjectID {
	return bson.NewObjectID()
}

func IsNilObjectID(objectID string) bool {
	return len(objectID) == 0 || objectID == bson.NilObjectID.Hex()
}

func IsObjectIDProvided(objectID bson.ObjectID) bool {
	return !objectID.IsZero()
}
