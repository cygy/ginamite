package functions

import (
	"context"
	"errors"
	"fmt"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreateCollections creates the collections.
func (f *Functions) CreateCollections(ctx context.Context, collections []CollectionToCreate) {
	db := f.database()

	for _, collection := range collections {
		// A nil builder must not be given to the driver: it would panic when listing its options.
		var opts []options.Lister[options.CreateCollectionOptions]
		if collection.Options != nil {
			opts = append(opts, collection.Options)
		}

		if err := db.CreateCollection(ctx, collection.Name, opts...); err != nil {
			log.Error("create mongoDB collection",
				"error", err.Error(),
				"collection", collection.Name,
			)
		}
	}
}

// RemoveCollections removes the collections.
func (f *Functions) RemoveCollections(ctx context.Context, collections []string) {
	db := f.database()

	for _, collection := range collections {
		if err := db.Collection(collection).Indexes().DropAll(ctx); err != nil {
			log.Error("unable to drop all", "error", err)
		}
		err := db.Collection(collection).Drop(ctx)
		if err != nil && !IsNotFound(err, errorCodeNamespaceNotFound, fmt.Sprintf("collection not found with name [%s]", collection)) {
			log.Error("delete mongoDB collection",
				"error", err.Error(),
				"collection", collection,
			)
		}
	}
}

// RemoveUnusedCollections removes the unused collections.
func (f *Functions) RemoveUnusedCollections(ctx context.Context) {
	collectionsToRemove := []string{}
	f.RemoveCollections(ctx, collectionsToRemove)
}

// CreateAllCollections creates the collections.
func (f *Functions) CreateAllCollections(ctx context.Context) {
	collectionOptions := options.CreateCollection().
		SetCollation(&options.Collation{
			Locale:          "fr",
			Strength:        1,
			NumericOrdering: false,
		})

	collections := []CollectionToCreate{
		{
			Name: model.AuthTokenCollection,
		},
		{
			Name:    model.DeletedUserCollection,
			Options: collectionOptions,
		},
		{
			Name:    model.DisabledUserCollection,
			Options: collectionOptions,
		},
		{
			Name: model.ForgotPasswordProcessCollection,
		},
		{
			Name: model.IPLocationCollection,
		},
		{
			Name: model.ProcessedMessageCollection,
		},
		{
			Name: model.UpdatePropertyProcessCollection,
		},
		{
			Name:    model.UserCollection,
			Options: collectionOptions,
		},
	}

	f.CreateCollections(ctx, collections)
}

const (
	// errorCodeNamespaceNotFound is the code of the MongoDB error returned for a collection that does not exist.
	errorCodeNamespaceNotFound = 26

	// errorCodeIndexNotFound is the code of the MongoDB error returned for an index that does not exist.
	errorCodeIndexNotFound = 27
)

// IsNotFound reports whether the error of the driver means that the collection or the index does not exist: it has
// the code of the server, or the message of the driver.
func IsNotFound(err error, code int, message string) bool {
	var commandError mongo.CommandError
	if errors.As(err, &commandError) && commandError.HasErrorCode(code) {
		return true
	}

	return err.Error() == message
}
