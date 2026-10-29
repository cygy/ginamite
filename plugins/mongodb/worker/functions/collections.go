package functions

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreateCollections : creates the collections.
func CreateCollections(ctx context.Context, collections []CollectionToCreate) {
	db := session.Database()

	for _, collection := range collections {
		err := db.CreateCollection(ctx, collection.Name, collection.Options)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error":      err.Error(),
				"collection": collection.Name,
			}).Error("create mongoDB collection")
		}
	}
}

// RemoveCollections : removes the collections.
func RemoveCollections(ctx context.Context, collections []string) {
	db := session.Database()

	for _, collection := range collections {
		db.Collection(collection).Indexes().DropAll(ctx)
		err := db.Collection(collection).Drop(ctx)
		if err != nil && err.Error() != fmt.Sprintf("collection not found with name [%s]", collection) {
			log.WithFields(logrus.Fields{
				"error":      err.Error(),
				"collection": collection,
			}).Error("delete mongoDB collection")
		}
	}
}

// RemoveUnusedCollections : removes the unused collections.
func RemoveUnusedCollections(ctx context.Context) {
	collectionsToRemove := []string{}
	RemoveCollections(ctx, collectionsToRemove)
}

// CreateAllCollections : creates the collections.
func CreateAllCollections(ctx context.Context) {
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
			Name: model.UpdatePropertyProcessCollection,
		},
		{
			Name:    model.UserCollection,
			Options: collectionOptions,
		},
	}

	CreateCollections(ctx, collections)
}
