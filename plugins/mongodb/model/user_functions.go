package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// IsUserWithUsernameExisting : returns true if a user with this username is already existing.
func IsUserWithUsernameExisting(username string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(UserCollection, bson.M{"username": username}, db)
	return count > 0
}

// IsUserWithEmailAddressExisting : returns true if a user with this email address is already existing.
func IsUserWithEmailAddressExisting(emailAddress string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(UserCollection, bson.M{"privateInfos.email": emailAddress}, db)
	return count > 0
}

// IsUserWithFacebookUserIDExisting : returns true if a user with this facebook user id is already existing.
func IsUserWithFacebookUserIDExisting(userID string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(UserCollection, bson.M{"identificationMethods.facebook.userId": bson.ObjectIdHex(userID)}, db)
	return count > 0
}

// IsUserWithGoogleUserIDExisting : returns true if a user with this google user id is already existing.
func IsUserWithGoogleUserIDExisting(userID string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(UserCollection, bson.M{"identificationMethods.google.userId": bson.ObjectIdHex(userID)}, db)
	return count > 0
}

// IsUserWithValidEmailAddressExisting : returns true if a user with this email address is already existing.
func IsUserWithValidEmailAddressExisting(emailAddress string, db *mgo.Database) bool {
	count, _ := document.CountBySelector(UserCollection, bson.M{"privateInfos.email": emailAddress, "privateInfos.isEmailValid": true}, db)
	return count > 0
}

// DeleteUserByUserID : deletes the document with the ID.
func DeleteUserByUserID(userID string, db *mgo.Database) error {
	return document.DeleteDocument(bson.ObjectIdHex(userID), UserCollection, db)
}

// GetAnonymousUserID : returns the ID of the anonymous user.
func GetAnonymousUserID(db *mgo.Database) (string, error) {
	user := User{}
	err := document.GetDocumentBySelector(&user, UserCollection, bson.M{"isAnonymous": true}, db)

	return user.ID.Hex(), err
}

// GetAnonymousUser : returns the anonymous user.
func GetAnonymousUser(db *mgo.Database) (*User, error) {
	user := &User{}
	err := document.GetDocumentBySelector(user, UserCollection, bson.M{"isAnonymous": true}, db)

	return user, err
}

// GetNeverUsedUserIDs : deletes the never used accounts from interval.
func GetNeverUsedUserIDs(interval time.Duration, db *mgo.Database) ([]string, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(&users, UserCollection, bson.D{{Name: "registrationInfos.date", Value: bson.M{"$lt": time.Now().Add(interval * -1)}}, {Name: "lastLogin", Value: bson.M{"$exists": false}}}, db)

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID.Hex()
	}

	return userIDs, err
}

// GetInactiveUserIDs : returns the ID of the inactive users.
func GetInactiveUserIDs(interval time.Duration, db *mgo.Database) ([]string, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(&users, UserCollection, bson.D{{Name: "lastLogin", Value: bson.M{"$lt": time.Now().Add(interval * -1)}}}, db)

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID.Hex()
	}

	return userIDs, err
}
