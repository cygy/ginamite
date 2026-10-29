package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// IsDisabledUserWithUsernameExisting : returns true if a disabled user with this username is already existing.
func IsDisabledUserWithUsernameExisting(ctx context.Context, username string, db *mongo.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"username": username}, db)
	return count > 0, err
}

// IsDisabledUserWithEmailAddressExisting : returns true if a disabled user with this email address is already existing.
func IsDisabledUserWithEmailAddressExisting(ctx context.Context, emailAddress string, db *mongo.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"privateInfos.email": emailAddress}, db)
	return count > 0, err
}

// IsDisabledUserWithFacebookUserIDExisting : returns true if a disabled user with this facebook user id is already existing.
func IsDisabledUserWithFacebookUserIDExisting(ctx context.Context, userID string, db *mongo.Database) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"identificationMethods.facebook.userId": objectID}, db)
	return count > 0, err
}

// IsDisabledUserWithGoogleUserIDExisting : returns true if a disabled user with this google user id is already existing.
func IsDisabledUserWithGoogleUserIDExisting(ctx context.Context, userID string, db *mongo.Database) (bool, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"identificationMethods.google.userId": objectID}, db)
	return count > 0, err
}

// IsDisabledUserWithValidEmailAddressExisting : returns true if a disabled user with this email address is already existing.
func IsDisabledUserWithValidEmailAddressExisting(ctx context.Context, emailAddress string, db *mongo.Database) (bool, error) {
	count, err := document.CountBySelector(ctx, DisabledUserCollection, bson.M{"privateInfos.email": emailAddress, "privateInfos.isEmailValid": true}, db)
	return count > 0, err
}

// DeleteDisabledUserByUserID : deletes the document with the ID.
func DeleteDisabledUserByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = document.DeleteDocumentsBySelector(ctx, DisabledUserCollection, bson.M{"userId": objectID}, db)
	return err
}

// GetDisabledUsersForAdmin : returns the disabled users for the admin.
func GetDisabledUsersForAdmin(ctx context.Context, db *mongo.Database) ([]AdminDisabledUser, error) {
	users := []AdminDisabledUser{}
	err := document.GetDocumentsBySort(ctx, &users, DisabledUserCollection, []string{"username"}, db)

	return users, err
}
