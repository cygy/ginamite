package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetByUsername : initializes the 'User' struct from the 'user' document with the username.
func (user *User) GetByUsername(ctx context.Context, username string, db *mongo.Database) error {
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"username": username}, db)
}

// GetByEmailAddress : initializes the 'User' struct from the 'user' document with the email address.
func (user *User) GetByEmailAddress(ctx context.Context, emailAddress string, db *mongo.Database) error {
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"privateInfos.email": emailAddress}, db)
}

// GetByFacebookUserID : initializes the 'User' struct from the 'user' document with the facebook userId.
func (user *User) GetByFacebookUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"identificationMethods.facebook.userId": userID}, db)
}

// GetByGoogleUserID : initializes the 'User' struct from the 'user' document with the google userId.
func (user *User) GetByGoogleUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"identificationMethods.google.userId": userID}, db)
}

// GetByIDAndNotification : initializes the 'User' struct from the 'user' document with the email address and a notification type.
func (user *User) GetByIDAndNotification(ctx context.Context, userID, notificationType string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return errors.WrongObjectID()
	}
	return document.GetDocumentBySelector(ctx, user, UserCollection, bson.D{
		{Key: "_id", Value: objectID},
		{Key: "notifications", Value: bson.M{"$in": []string{notificationType}}},
	}, db)
}

// GetByIdentifier : initializes the 'User' struct from the 'user' document with the username or the email address.
func (user *User) GetByIdentifier(ctx context.Context, identifier string, db *mongo.Database) error {
	if err := document.GetDocumentBySelector(ctx, user, UserCollection, bson.D{{Key: "username", Value: identifier}}, db); err != nil && !errors.IsNotFound(err) {
		return err
	}

	if !user.IsSaved() {
		if err := document.GetDocumentBySelector(ctx, user, UserCollection, bson.D{{Key: "privateInfos.email", Value: identifier}}, db); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	if !user.IsSaved() {
		return errors.NotFound()
	}

	return nil
}

// UpdateLastLogin : updates the last login date.
func (user *User) UpdateLastLogin(ctx context.Context, db *mongo.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"lastLogin": time.Now()}}, db)
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
