package functions

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// IndexToCreate : struct representing the index of a collection.
type IndexToCreate struct {
	Collection string
	Index      mongo.IndexModel
}

// IndexToDelete : struct representing the index of a collection.
type IndexToDelete struct {
	Collection string
	IndexName  string
}

// CollectionToCreate : struct representing a collection.
type CollectionToCreate struct {
	Name    string
	Options *options.CreateCollectionOptionsBuilder
}
