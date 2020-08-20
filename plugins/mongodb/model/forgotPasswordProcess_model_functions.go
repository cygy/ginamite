package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetByUserID : returns a ForgotPasswordProcess struct with a user.
func (fpp *ForgotPasswordProcess) GetByUserID(userID string, db *mgo.Database) error {
	return document.GetDocumentBySelector(fpp, ForgotPasswordProcessCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, db)
}

// Increment : increments the tries of a forgot password proccess document.
func (fpp *ForgotPasswordProcess) Increment(db *mgo.Database) error {
	return fpp.Update(bson.M{"$inc": bson.M{"tries": 1}}, db)
}
