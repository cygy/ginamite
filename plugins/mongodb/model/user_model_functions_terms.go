package model

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// AcceptVersionOfTerms : updates the version of the terms accepted.
func (user *User) AcceptVersionOfTerms(version string, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"versionOfTermsAccepted": version}}, db)
}
