package model

import (
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// DeleteUpdatePropertyProcessByUserIDAndType : deletes a process linked to a user.
func DeleteUpdatePropertyProcessByUserIDAndType(userID, t string, db *mgo.Database) error {
	_, err := document.DeleteDocumentsBySelector(UpdatePropertyProcessCollection, bson.D{{Name: "userId", Value: bson.ObjectIdHex(userID)}, {Name: "type", Value: t}}, db)
	return err
}

// DeleteUpdatePasswordProcessByUserID : deletes a process linked to a user.
func DeleteUpdatePasswordProcessByUserID(userID string, db *mgo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(userID, authentication.ProcessUpdatePassword, db)
}

// DeleteUpdateEmailAddressProcessByUserID : deletes a process linked to a user.
func DeleteUpdateEmailAddressProcessByUserID(userID string, db *mgo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(userID, authentication.ProcessUpdateEmailAddress, db)
}

// DeleteUpdateValidEmailAddressProcessByUserID : deletes a process linked to a user.
func DeleteUpdateValidEmailAddressProcessByUserID(userID string, db *mgo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(userID, authentication.ProcessVerifyEmailAddress, db)
}

// DeleteDeleteAccountProcessByUserID : deletes a process linked to a user.
func DeleteDeleteAccountProcessByUserID(userID string, db *mgo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(userID, authentication.ProcessDeleteAccount, db)
}

// DeleteDisableAccountProcessByUserID : deletes a process linked to a user.
func DeleteDisableAccountProcessByUserID(userID string, db *mgo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(userID, authentication.ProcessDisableAccount, db)
}
