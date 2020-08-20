package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// DeleteDeletedUserByUserID : deletes the document with the ID.
func DeleteDeletedUserByUserID(userID string, db *mgo.Database) error {
	_, err := document.DeleteDocumentsBySelector(DeletedUserCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, db)
	return err
}
