package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// DeleteForgotPasswordProcessByUserID : deletes all the documents linked to a user.
func DeleteForgotPasswordProcessByUserID(userID string, db *mgo.Database) error {
	_, err := document.DeleteDocumentsBySelector(ForgotPasswordProcessCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, db)
	return err
}
