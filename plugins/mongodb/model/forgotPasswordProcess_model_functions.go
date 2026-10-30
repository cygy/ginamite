package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// GetByUserID : returns a ForgotPasswordProcess struct with a user.
func (fpp *ForgotPasswordProcess) GetByUserID(ctx context.Context, userID string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return document.GetDocumentBySelector(ctx, fpp, ForgotPasswordProcessCollection, bson.M{"userId": objectID}, db)
}

// Increment : increments the tries of a forgot password proccess document.
func (fpp *ForgotPasswordProcess) Increment(ctx context.Context, db store.Database) error {
	return fpp.Update(ctx, bson.M{"$inc": bson.M{"tries": 1}}, db)
}
