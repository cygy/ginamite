package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// DeleteDeletedUserByUserID : deletes the document with the ID.
func DeleteDeletedUserByUserID(ctx context.Context, userID string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = document.DeleteDocumentsBySelector(ctx, DeletedUserCollection, bson.M{"userId": objectID}, db)
	return err
}
