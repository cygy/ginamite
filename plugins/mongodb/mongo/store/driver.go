package store

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NewDatabase returns the Database backed by the driver database db, or nil if db is nil.
func NewDatabase(db *mongo.Database) Database {
	if db == nil {
		return nil
	}

	return driverDatabase{db: db}
}

// driverDatabase adapts a *mongo.Database to Database.
type driverDatabase struct {
	db *mongo.Database
}

// Name returns the name of the database.
func (d driverDatabase) Name() string {
	return d.db.Name()
}

// Collection returns the collection with the name.
func (d driverDatabase) Collection(name string) Collection {
	return driverCollection{coll: d.db.Collection(name)}
}

// CreateCollection creates a collection explicitly.
func (d driverDatabase) CreateCollection(ctx context.Context, name string, opts ...options.Lister[options.CreateCollectionOptions]) error {
	return d.db.CreateCollection(ctx, name, opts...)
}

// matchAll returns an empty filter for a nil filter, which mgo accepted to match all the documents and which the
// driver rejects with mongo.ErrNilDocument.
func matchAll(filter any) any {
	if filter == nil {
		return bson.D{}
	}

	// A nil bson.D, bson.M or pointer in an interface is not a nil interface, and is marshaled as null.
	switch value := reflect.ValueOf(filter); value.Kind() {
	case reflect.Slice, reflect.Map, reflect.Pointer:
		if value.IsNil() {
			return bson.D{}
		}
	}

	return filter
}

// driverCollection adapts a *mongo.Collection to Collection.
type driverCollection struct {
	coll *mongo.Collection
}

func (c driverCollection) InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	return c.coll.InsertOne(ctx, document, opts...)
}

func (c driverCollection) UpdateByID(ctx context.Context, id, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	return c.coll.UpdateByID(ctx, id, update, opts...)
}

func (c driverCollection) UpdateOne(ctx context.Context, filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	return c.coll.UpdateOne(ctx, matchAll(filter), update, opts...)
}

func (c driverCollection) UpdateMany(ctx context.Context, filter, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	return c.coll.UpdateMany(ctx, matchAll(filter), update, opts...)
}

func (c driverCollection) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	return c.coll.DeleteOne(ctx, matchAll(filter), opts...)
}

func (c driverCollection) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	return c.coll.DeleteMany(ctx, matchAll(filter), opts...)
}

func (c driverCollection) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResult {
	return c.coll.FindOne(ctx, matchAll(filter), opts...)
}

func (c driverCollection) FindOneAndUpdate(ctx context.Context, filter, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) SingleResult {
	return c.coll.FindOneAndUpdate(ctx, matchAll(filter), update, opts...)
}

// Find returns a nil Cursor, not a typed nil, if the query fails.
func (c driverCollection) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (Cursor, error) {
	cursor, err := c.coll.Find(ctx, matchAll(filter), opts...)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

func (c driverCollection) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return c.coll.CountDocuments(ctx, matchAll(filter), opts...)
}

// Aggregate returns a nil Cursor, not a typed nil, if the query fails.
func (c driverCollection) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (Cursor, error) {
	cursor, err := c.coll.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

func (c driverCollection) Indexes() IndexView {
	return c.coll.Indexes()
}

func (c driverCollection) Drop(ctx context.Context, opts ...options.Lister[options.DropCollectionOptions]) error {
	return c.coll.Drop(ctx, opts...)
}
