package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// DeleteForgotPasswordProcessByUserID : deletes all the documents linked to a user.
func DeleteForgotPasswordProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = document.DeleteDocumentsBySelector(ctx, ForgotPasswordProcessCollection, bson.M{"userId": objectID}, db)
	return err
}
