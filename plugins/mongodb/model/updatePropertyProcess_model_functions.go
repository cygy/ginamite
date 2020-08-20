package model

import (
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetByUserIDAndType : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetByUserIDAndType(userID, t string, db *mgo.Database) error {
	return document.GetDocumentBySelector(process, UpdatePropertyProcessCollection, bson.D{{Name: "userId", Value: bson.ObjectIdHex(userID)}, {Name: "type", Value: t}}, db)
}

// GetUpdatePasswordProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdatePasswordProcessByUserID(userID string, db *mgo.Database) error {
	return process.GetByUserIDAndType(userID, authentication.ProcessUpdatePassword, db)
}

// GetUpdateEmailAddressProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdateEmailAddressProcessByUserID(userID string, db *mgo.Database) error {
	return process.GetByUserIDAndType(userID, authentication.ProcessUpdateEmailAddress, db)
}

// GetUpdateValidEmailAddressProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdateValidEmailAddressProcessByUserID(userID string, db *mgo.Database) error {
	return process.GetByUserIDAndType(userID, authentication.ProcessVerifyEmailAddress, db)
}

// GetDeleteAccountProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetDeleteAccountProcessByUserID(userID string, db *mgo.Database) error {
	return process.GetByUserIDAndType(userID, authentication.ProcessDeleteAccount, db)
}

// GetDisableAccountProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetDisableAccountProcessByUserID(userID string, db *mgo.Database) error {
	return process.GetByUserIDAndType(userID, authentication.ProcessDisableAccount, db)
}

// Increment : increments the tries of a UpdatePropertyProcess document.
func (process *UpdatePropertyProcess) Increment(db *mgo.Database) error {
	return process.Update(bson.M{"$inc": bson.M{"tries": 1}}, db)
}
