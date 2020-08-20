package model

import (
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetByUsername : initializes the 'User' struct from the 'user' document with the username.
func (user *User) GetByUsername(username string, db *mgo.Database) error {
	return document.GetDocumentBySelector(user, UserCollection, bson.M{"username": username}, db)
}

// GetByEmailAddress : initializes the 'User' struct from the 'user' document with the email address.
func (user *User) GetByEmailAddress(emailAddress string, db *mgo.Database) error {
	return document.GetDocumentBySelector(user, UserCollection, bson.M{"privateInfos.email": emailAddress}, db)
}

// GetByFacebookUserID : initializes the 'User' struct from the 'user' document with the facebook userId.
func (user *User) GetByFacebookUserID(userID string, db *mgo.Database) error {
	return document.GetDocumentBySelector(user, UserCollection, bson.M{"identificationMethods.facebook.userId": userID}, db)
}

// GetByGoogleUserID : initializes the 'User' struct from the 'user' document with the google userId.
func (user *User) GetByGoogleUserID(userID string, db *mgo.Database) error {
	return document.GetDocumentBySelector(user, UserCollection, bson.M{"identificationMethods.google.userId": userID}, db)
}

// GetByIDAndNotification : initializes the 'User' struct from the 'user' document with the email address and a notification type.
func (user *User) GetByIDAndNotification(userID, notificationType string, db *mgo.Database) error {
	return document.GetDocumentBySelector(user, UserCollection, bson.D{{Name: "_id", Value: bson.ObjectIdHex(userID)}, {Name: "notifications", Value: bson.M{"$in": []string{notificationType}}}}, db)
}

// GetByIdentifier : initializes the 'User' struct from the 'user' document with the username or the email address.
func (user *User) GetByIdentifier(identifier string, db *mgo.Database) error {
	if err := document.GetDocumentBySelector(user, UserCollection, bson.D{{Name: "username", Value: identifier}}, db); err != nil && !errors.IsNotFound(err) {
		return err
	}

	if !user.IsSaved() {
		if err := document.GetDocumentBySelector(user, UserCollection, bson.D{{Name: "privateInfos.email", Value: identifier}}, db); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	if !user.IsSaved() {
		return errors.NotFound()
	}

	return nil
}

// UpdateLastLogin : updates the last login date.
func (user *User) UpdateLastLogin(db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"lastLogin": time.Now()}}, db)
}

// ToUser : return a authentication.User struct from this model.User struct.
func (user *User) ToUser() authentication.User {
	return authentication.User{
		ID:           user.ID.Hex(),
		Username:     user.Username,
		EmailAddress: user.PrivateInfos.Email,
		Locale:       user.Settings.Locale,
		PrivateKey:   user.PrivateKey,
		IsEmailValid: user.PrivateInfos.IsEmailValid,
	}
}
