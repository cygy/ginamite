package model

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UpdateRead : updated the read property of a notification received by a user.
func (notification *Notification) UpdateRead(db *mgo.Database) error {
	return notification.Update(bson.M{"$set": bson.M{"read": true}}, db)
}

// UpdateNotified : updated the notified property of a notification received by a user.
func (notification *Notification) UpdateNotified(db *mgo.Database) error {
	return notification.Update(bson.M{"$set": bson.M{"notified": true}}, db)
}
