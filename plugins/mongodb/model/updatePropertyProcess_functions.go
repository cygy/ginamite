package model

import (
	"context"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// DeleteUpdatePropertyProcessByUserIDAndType : deletes a process linked to a user.
func DeleteUpdatePropertyProcessByUserIDAndType(ctx context.Context, userID, t string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return errors.WrongObjectID()
	}
	_, err = document.DeleteDocumentsBySelector(ctx, UpdatePropertyProcessCollection, bson.D{
		{Key: "userId", Value: objectID},
		{Key: "type", Value: t},
	}, db)
	return err
}

// DeleteUpdatePasswordProcessByUserID : deletes a process linked to a user.
func DeleteUpdatePasswordProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessUpdatePassword, db)
}

// DeleteUpdateEmailAddressProcessByUserID : deletes a process linked to a user.
func DeleteUpdateEmailAddressProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessUpdateEmailAddress, db)
}

// DeleteUpdateValidEmailAddressProcessByUserID : deletes a process linked to a user.
func DeleteUpdateValidEmailAddressProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessVerifyEmailAddress, db)
}

// DeleteDeleteAccountProcessByUserID : deletes a process linked to a user.
func DeleteDeleteAccountProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessDeleteAccount, db)
}

// DeleteDisableAccountProcessByUserID : deletes a process linked to a user.
func DeleteDisableAccountProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return DeleteUpdatePropertyProcessByUserIDAndType(ctx, userID, authentication.ProcessDisableAccount, db)
}
