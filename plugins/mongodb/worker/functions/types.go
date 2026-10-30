// Package functions implements the functions run by the workers on the data stored in MongoDB: the creation of the
// collections and of the indexes, the maintenance of the accounts and the handling of the tasks.
package functions

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// IndexToCreate is an index to create in a collection.
type IndexToCreate struct {
	Collection string
	Index      mongo.IndexModel
}

// IndexToDelete is an index to delete from a collection, by its name.
type IndexToDelete struct {
	Collection string
	IndexName  string
}

// CollectionToCreate is a collection to create, with its options if any.
type CollectionToCreate struct {
	Name    string
	Options *options.CreateCollectionOptionsBuilder
}
