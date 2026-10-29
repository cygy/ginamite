package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// IsUserWithUsernameExisting : returns true if a user with this username is already existing.
func IsUserWithUsernameExisting(ctx context.Context, username string, db *mongo.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"username": username}, db)
	return count > 0, err
}

// IsUserWithEmailAddressExisting : returns true if a user with this email address is already existing.
func IsUserWithEmailAddressExisting(ctx context.Context, emailAddress string, db *mongo.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"privateInfos.email": emailAddress}, db)
	return count > 0, err
}

// IsUserWithFacebookUserIDExisting : returns true if a user with this facebook user id is already existing.
func IsUserWithFacebookUserIDExisting(ctx context.Context, userID string, db *mongo.Database) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"identificationMethods.facebook.userId": objectID}, db)
	return count > 0, err
}

// IsUserWithGoogleUserIDExisting : returns true if a user with this google user id is already existing.
func IsUserWithGoogleUserIDExisting(ctx context.Context, userID string, db *mongo.Database) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"identificationMethods.google.userId": objectID}, db)
	return count > 0, err
}

// IsUserWithValidEmailAddressExisting : returns true if a user with this email address is already existing.
func IsUserWithValidEmailAddressExisting(ctx context.Context, emailAddress string, db *mongo.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, UserCollection, bson.M{"privateInfos.email": emailAddress, "privateInfos.isEmailValid": true}, db)
	return count > 0, err
}

// DeleteUserByUserID : deletes the document with the ID.
func DeleteUserByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return document.DeleteDocument(ctx, objectID, UserCollection, db)
}

// GetAnonymousUserID : returns the ID of the anonymous user.
func GetAnonymousUserID(ctx context.Context, db *mongo.Database) (string, error) {
	user := User{}
	err := document.GetDocumentBySelector(ctx, &user, UserCollection, bson.M{"isAnonymous": true}, db)
	return user.ID.Hex(), err
}

// GetAnonymousUser : returns the anonymous user.
func GetAnonymousUser(ctx context.Context, db *mongo.Database) (*User, error) {
	user := &User{}
	err := document.GetDocumentBySelector(ctx, user, UserCollection, bson.M{"isAnonymous": true}, db)
	return user, err
}

// GetNeverUsedUserIDs : returns the user that has never been used.
func GetNeverUsedUserIDs(ctx context.Context, interval time.Duration, db *mongo.Database) ([]string, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(ctx, &users, UserCollection,
		bson.D{
			{Key: "registrationInfos.date", Value: bson.M{"$lt": time.Now().Add(interval * -1)}},
			{Key: "lastLogin", Value: bson.M{"$exists": false}},
			{Key: "isAdmin", Value: false},
		},
		db)

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID.Hex()
	}

	return userIDs, err
}

// GetInactiveUserIDs : returns the ID of the inactive users.
func GetInactiveUserIDs(ctx context.Context, interval time.Duration, db *mongo.Database) ([]string, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(ctx, &users, UserCollection, bson.D{{Key: "lastLogin", Value: bson.M{"$lt": time.Now().Add(interval * -1)}}}, db)

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID.Hex()
	}

	return userIDs, err
}
