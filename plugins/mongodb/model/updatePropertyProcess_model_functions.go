package model

import (
	"context"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetByUserIDAndType : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetByUserIDAndType(ctx context.Context, userID, t string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return document.GetDocumentBySelector(ctx, process, UpdatePropertyProcessCollection, bson.D{
		{Key: "userId", Value: objectID},
		{Key: "type", Value: t},
	}, db)
}

// GetUpdatePasswordProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdatePasswordProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessUpdatePassword, db)
}

// GetUpdateEmailAddressProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdateEmailAddressProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessUpdateEmailAddress, db)
}

// GetUpdateValidEmailAddressProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetUpdateValidEmailAddressProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessVerifyEmailAddress, db)
}

// GetDeleteAccountProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetDeleteAccountProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessDeleteAccount, db)
}

// GetDisableAccountProcessByUserID : returns a UpdatePropertyProcess struct with a user.
func (process *UpdatePropertyProcess) GetDisableAccountProcessByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	return process.GetByUserIDAndType(ctx, userID, authentication.ProcessDisableAccount, db)
}

// Increment : increments the tries of a UpdatePropertyProcess document.
func (process *UpdatePropertyProcess) Increment(ctx context.Context, db *mongo.Database) error {
	return process.Update(ctx, bson.M{"$inc": bson.M{"tries": 1}}, db)
}
