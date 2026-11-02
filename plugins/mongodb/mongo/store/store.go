// Package store defines the subset of the MongoDB driver used by the application as interfaces,
// so that the code accessing the database can be tested with a stub, see the mock package.
// Use NewDatabase to get a Database backed by the driver.
package store

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Database is a MongoDB database.
type Database interface {
	// Name returns the name of the database.
	Name() string

	// Collection returns the collection with the name.
	Collection(name string) Collection

	// CreateCollection creates a collection explicitly.
	CreateCollection(ctx context.Context, name string, opts ...options.Lister[options.CreateCollectionOptions]) error
}

// Collection is a MongoDB collection.
type Collection interface {
	InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error)
	UpdateByID(ctx context.Context, id, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error)
	UpdateOne(ctx context.Context, filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error)
	FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResult
	FindOneAndUpdate(ctx context.Context, filter, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) SingleResult
	Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (Cursor, error)
	CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error)
	Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (Cursor, error)

	// Indexes returns the indexes of the collection.
	Indexes() IndexView

	// Drop removes the collection from the database.
	Drop(ctx context.Context, opts ...options.Lister[options.DropCollectionOptions]) error
}

// SingleResult is the result of a query returning at most one document.
type SingleResult interface {
	// Decode decodes the document into v. It returns mongo.ErrNoDocuments if there is none.
	Decode(v any) error

	// Err returns the error of the query, if any.
	Err() error
}

// Cursor iterates over the documents returned by a query.
type Cursor interface {
	// All decodes all the remaining documents into a pointer to a slice, and closes the cursor.
	All(ctx context.Context, results any) error

	// Next advances to the next document and reports whether there is one.
	Next(ctx context.Context) bool

	// Decode decodes the current document into v.
	Decode(v any) error

	// Err returns the error that stopped the iteration, if any.
	Err() error

	// Close releases the resources of the cursor.
	Close(ctx context.Context) error
}

// IndexView manages the indexes of a collection.
type IndexView interface {
	CreateOne(ctx context.Context, model mongo.IndexModel, opts ...options.Lister[options.CreateIndexesOptions]) (string, error)
	DropOne(ctx context.Context, name string, opts ...options.Lister[options.DropIndexesOptions]) error
	DropAll(ctx context.Context, opts ...options.Lister[options.DropIndexesOptions]) error
}
