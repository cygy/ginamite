package functions

import (
	"fmt"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/globalsign/mgo"
	"github.com/sirupsen/logrus"
)

// CreateCollections : creates the collections.
func CreateCollections(collections []CollectionToCreate) {
	session, db := session.Copy()
	defer session.Close()

	for _, collection := range collections {
		err := db.C(collection.Name).Create(&collection.Info)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error":      err.Error(),
				"collection": collection.Name,
			}).Error("create mongoDB collection")
		}
	}
}

// RemoveCollections : removes the collections.
func RemoveCollections(collections []string) {
	session, db := session.Copy()
	defer session.Close()

	for _, collection := range collections {
		db.C(collection).DropAllIndexes()
		err := db.C(collection).DropCollection()
		if err != nil && err.Error() != fmt.Sprintf("collection not found with name [%s]", collection) {
			log.WithFields(logrus.Fields{
				"error":      err.Error(),
				"collection": collection,
			}).Error("delete mongoDB collection")
		}
	}
}

// RemoveUnusedCollections : removes the unused collections.
func RemoveUnusedCollections() {
	collectionsToRemove := []string{}

	RemoveCollections(collectionsToRemove)
}

// CreateAllCollections : creates the collections.
func CreateAllCollections() {
	collectionInfo := mgo.CollectionInfo{
		Collation: &mgo.Collation{
			Locale:          "fr",
			Strength:        1,
			NumericOrdering: false,
		},
	}

	collections := []CollectionToCreate{
		{
			Name: model.AuthTokenCollection,
		},
		{
			Name: model.DeletedUserCollection,
			Info: collectionInfo,
		},
		{
			Name: model.DisabledUserCollection,
			Info: collectionInfo,
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
			Name: model.UserCollection,
			Info: collectionInfo,
		},
	}

	CreateCollections(collections)
}
