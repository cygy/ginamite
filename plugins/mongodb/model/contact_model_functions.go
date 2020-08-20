package model

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UpdateDone : updates the 'done' properties of a contact.
func (contact *Contact) UpdateDone(done bool, db *mgo.Database) error {
	return contact.Update(bson.M{"$set": bson.M{"done": done}}, db)
}
