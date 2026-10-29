package model

import (
	"context"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// GetByUserID : returns a ForgotPasswordProcess struct with a user.
func (fpp *ForgotPasswordProcess) GetByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return errors.WrongObjectID()
	}
	return document.GetDocumentBySelector(ctx, fpp, ForgotPasswordProcessCollection, bson.M{"userId": objectID}, db)
}

// Increment : increments the tries of a forgot password proccess document.
func (fpp *ForgotPasswordProcess) Increment(ctx context.Context, db *mongo.Database) error {
	return fpp.Update(ctx, bson.M{"$inc": bson.M{"tries": 1}}, db)
}
