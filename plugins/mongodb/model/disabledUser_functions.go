package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// IsDisabledUserWithUsernameExisting : returns true if a disabled user with this username is already existing.
func IsDisabledUserWithUsernameExisting(username string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(DisabledUserCollection, bson.M{"username": username}, db)
	return count > 0
}

// IsDisabledUserWithEmailAddressExisting : returns true if a disabled user with this email address is already existing.
func IsDisabledUserWithEmailAddressExisting(emailAddress string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(DisabledUserCollection, bson.M{"privateInfos.email": emailAddress}, db)
	return count > 0
}

// IsDisabledUserWithFacebookUserIDExisting : returns true if a disabled user with this facebook user id is already existing.
func IsDisabledUserWithFacebookUserIDExisting(userID string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(DisabledUserCollection, bson.M{"identificationMethods.facebook.userId": bson.ObjectIdHex(userID)}, db)
	return count > 0
}

// IsDisabledUserWithGoogleUserIDExisting : returns true if a disabled user with this google user id is already existing.
func IsDisabledUserWithGoogleUserIDExisting(userID string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(DisabledUserCollection, bson.M{"identificationMethods.google.userId": bson.ObjectIdHex(userID)}, db)
	return count > 0
}

// IsDisabledUserWithValidEmailAddressExisting : returns true if a disabled user with this email address is already existing.
func IsDisabledUserWithValidEmailAddressExisting(emailAddress string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(DisabledUserCollection, bson.M{"privateInfos.email": emailAddress, "privateInfos.isEmailValid": true}, db)
	return count > 0
}

// DeleteDisabledUserByUserID : deletes the document with the ID.
func DeleteDisabledUserByUserID(userID string, db *mgo.Database) error {
	_, err := document.DeleteDocumentsBySelector(DisabledUserCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, db)
	return err
}

// GetDisabledUsersForAdmin : returns the disabled users for the admin.
func GetDisabledUsersForAdmin(db *mgo.Database) ([]AdminDisabledUser, error) {
	users := []AdminDisabledUser{}
	err := document.GetDocumentsBySort(&users, DisabledUserCollection, []string{"username"}, db)

	return users, err
}
