// Package mock provides a stub of the store interfaces to test the code accessing MongoDB without a server.
// The stub does not evaluate the filters, the updates or the pipelines: it records the calls of each collection
// and returns the results and the errors configured by the test.
package mock

import (
	"context"
	"reflect"
	"sync"

	"github.com/cygy/ginamite/common/mongo/store"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Call is a call of a collection method.
type Call struct {
	// Method is the name of the method, e.g. "FindOne".
	Method string

	// Args are the arguments of the call, without the context and the options.
	Args []any

	// Opts are the options of the call.
	Opts []any
}

// Database is a stub of store.Database.
type Database struct {
	// DatabaseName is returned by Name.
	DatabaseName string

	// CreateCollectionErr is returned by CreateCollection.
	CreateCollectionErr error

	mu          sync.Mutex
	collections map[string]*Collection
	created     []string
	createOpts  map[string][]options.Lister[options.CreateCollectionOptions]
}

// NewDatabase returns a stub database with the name.
func NewDatabase(name string) *Database {
	return &Database{DatabaseName: name, collections: map[string]*Collection{}}
}

// Name returns the name of the database.
func (d *Database) Name() string {
	return d.DatabaseName
}

// Collection returns the stub collection with the name, created at the first call.
func (d *Database) Collection(name string) store.Collection {
	return d.Coll(name)
}

// Coll returns the stub collection with the name, created at the first call, to configure it or to read its calls.
func (d *Database) Coll(name string) *Collection {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.collections == nil {
		d.collections = map[string]*Collection{}
	}

	c, ok := d.collections[name]
	if !ok {
		c = &Collection{Name: name}
		d.collections[name] = c
	}

	return c
}

// CreateCollection records the name and the options of the collection and returns CreateCollectionErr.
func (d *Database) CreateCollection(_ context.Context, name string, opts ...options.Lister[options.CreateCollectionOptions]) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.created = append(d.created, name)
	if d.createOpts == nil {
		d.createOpts = map[string][]options.Lister[options.CreateCollectionOptions]{}
	}
	d.createOpts[name] = opts

	return d.CreateCollectionErr
}

// CreateOptions returns the options given when the collection was created.
func (d *Database) CreateOptions(name string) []options.Lister[options.CreateCollectionOptions] {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.createOpts[name]
}

// Created returns the names of the collections created with CreateCollection.
func (d *Database) Created() []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	return append([]string(nil), d.created...)
}

// Collection is a stub of store.Collection.
//
// A method returns the value of the matching field. The zero values are valid: a write reports one
// matched, modified or deleted document, a query finds no document and a count is zero.
type Collection struct {
	// Name is the name of the collection.
	Name string

	// Err is returned by every method that returns an error, unless the method has its own error in Errs.
	Err error

	// Errs are the errors returned by a given method, e.g. "CountDocuments".
	Errs map[string]error

	// InsertResult, UpdateResult and DeleteResult are returned by the writes. They default to one affected document.
	InsertResult *mongo.InsertOneResult
	UpdateResult *mongo.UpdateResult
	DeleteResult *mongo.DeleteResult

	// One is the result of FindOne. It defaults to a result without document (mongo.ErrNoDocuments).
	One store.SingleResult

	// Docs are the documents returned by Find and Aggregate, in order.
	Docs []any

	// CloseErr is returned when closing the cursors returned by Find and Aggregate.
	CloseErr error

	// Count is returned by CountDocuments.
	Count int64

	// IndexView is returned by Indexes. It defaults to a stub recording its calls.
	IndexView *IndexView

	mu    sync.Mutex
	calls []Call
}

// record saves the call and returns the error to return for the method, if any.
func (c *Collection) record(method string, args []any, opts []any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.calls = append(c.calls, Call{Method: method, Args: args, Opts: opts})

	if err, ok := c.Errs[method]; ok {
		return err
	}

	return c.Err
}

// Calls returns the calls received by the collection, in order.
func (c *Collection) Calls() []Call {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]Call(nil), c.calls...)
}

// LastCall returns the last call received by the collection, or a zero Call if there is none.
func (c *Collection) LastCall() Call {
	calls := c.Calls()
	if len(calls) == 0 {
		return Call{}
	}

	return calls[len(calls)-1]
}

func listerArgs[T any](opts []options.Lister[T]) []any {
	args := make([]any, len(opts))
	for i, o := range opts {
		args[i] = o
	}

	return args
}

// InsertOne records the call.
func (c *Collection) InsertOne(_ context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	if err := c.record("InsertOne", []any{document}, listerArgs(opts)); err != nil {
		return nil, err
	}
	if c.InsertResult != nil {
		return c.InsertResult, nil
	}

	return &mongo.InsertOneResult{}, nil
}

func (c *Collection) updated(err error) (*mongo.UpdateResult, error) {
	if err != nil {
		return nil, err
	}
	if c.UpdateResult != nil {
		return c.UpdateResult, nil
	}

	return &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
}

// UpdateByID records the call.
func (c *Collection) UpdateByID(_ context.Context, id, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	return c.updated(c.record("UpdateByID", []any{id, update}, listerArgs(opts)))
}

// UpdateOne records the call.
func (c *Collection) UpdateOne(_ context.Context, filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	return c.updated(c.record("UpdateOne", []any{filter, update}, listerArgs(opts)))
}

// UpdateMany records the call.
func (c *Collection) UpdateMany(_ context.Context, filter, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	return c.updated(c.record("UpdateMany", []any{filter, update}, listerArgs(opts)))
}

func (c *Collection) deleted(err error) (*mongo.DeleteResult, error) {
	if err != nil {
		return nil, err
	}
	if c.DeleteResult != nil {
		return c.DeleteResult, nil
	}

	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

// DeleteOne records the call.
func (c *Collection) DeleteOne(_ context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	return c.deleted(c.record("DeleteOne", []any{filter}, listerArgs(opts)))
}

// DeleteMany records the call.
func (c *Collection) DeleteMany(_ context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	return c.deleted(c.record("DeleteMany", []any{filter}, listerArgs(opts)))
}

// FindOne records the call and returns One, or Err, or a result without document.
func (c *Collection) FindOne(_ context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) store.SingleResult {
	if err := c.record("FindOne", []any{filter}, listerArgs(opts)); err != nil {
		return Result{Error: err}
	}
	if c.One != nil {
		return c.One
	}

	return Result{Error: mongo.ErrNoDocuments}
}

// Find records the call and returns a cursor over Docs, or Err.
func (c *Collection) Find(_ context.Context, filter any, opts ...options.Lister[options.FindOptions]) (store.Cursor, error) {
	if err := c.record("Find", []any{filter}, listerArgs(opts)); err != nil {
		return nil, err
	}

	return c.cursor(), nil
}

// cursor returns a cursor over Docs.
func (c *Collection) cursor() *Cursor {
	cursor := NewCursor(c.Docs...)
	cursor.CloseErr = c.CloseErr

	return cursor
}

// CountDocuments records the call and returns Count, or Err.
func (c *Collection) CountDocuments(_ context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	if err := c.record("CountDocuments", []any{filter}, listerArgs(opts)); err != nil {
		return 0, err
	}

	return c.Count, nil
}

// Aggregate records the call and returns a cursor over Docs, or Err.
func (c *Collection) Aggregate(_ context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (store.Cursor, error) {
	if err := c.record("Aggregate", []any{pipeline}, listerArgs(opts)); err != nil {
		return nil, err
	}

	return c.cursor(), nil
}

// Indexes returns IndexView, created at the first call.
func (c *Collection) Indexes() store.IndexView {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IndexView == nil {
		c.IndexView = &IndexView{}
	}

	return c.IndexView
}

// Drop records the call and returns Err.
func (c *Collection) Drop(_ context.Context, _ ...options.Lister[options.DropCollectionOptions]) error {
	return c.record("Drop", nil, nil)
}

// applyOptions applies the options of a call to a new T.
func applyOptions[T any](opts []any) *T {
	args := new(T)
	for _, o := range opts {
		if lister, ok := o.(options.Lister[T]); ok {
			for _, set := range lister.List() {
				_ = set(args)
			}
		}
	}

	return args
}

// FindOptions applies the options of a Find call, e.g. to check the sort, the skip and the limit.
func FindOptions(opts []any) *options.FindOptions {
	return applyOptions[options.FindOptions](opts)
}

// FindOneOptions applies the options of a FindOne call, e.g. to check the sort.
func FindOneOptions(opts []any) *options.FindOneOptions {
	return applyOptions[options.FindOneOptions](opts)
}

// UpdateOneOptions applies the options of an UpdateOne call, e.g. to check the upsert.
func UpdateOneOptions(opts []any) *options.UpdateOneOptions {
	return applyOptions[options.UpdateOneOptions](opts)
}

// Result is a stub of store.SingleResult.
type Result struct {
	// Value is the document returned by Decode.
	Value any

	// Error is returned by Err and Decode. Decode returns mongo.ErrNoDocuments if both are nil and Value is nil.
	Error error
}

// Decode copies the value into v, which must be a pointer, by way of BSON, like the driver does.
func (r Result) Decode(v any) error {
	if r.Error != nil {
		return r.Error
	}
	if r.Value == nil {
		return mongo.ErrNoDocuments
	}

	return convert(r.Value, v)
}

// Err returns the error of the result.
func (r Result) Err() error {
	return r.Error
}

// Cursor is a stub of store.Cursor over a list of documents.
type Cursor struct {
	// CloseErr is returned by Close.
	CloseErr error

	// DecodeErr is returned by Decode and All, when set.
	DecodeErr error

	docs   []any
	index  int
	closed bool
}

// NewCursor returns a cursor over the documents.
func NewCursor(docs ...any) *Cursor {
	return &Cursor{docs: docs, index: -1}
}

// Closed reports whether the cursor was closed.
func (c *Cursor) Closed() bool {
	return c.closed
}

// Next advances to the next document and reports whether there is one.
func (c *Cursor) Next(context.Context) bool {
	if c.index+1 >= len(c.docs) {
		return false
	}
	c.index++

	return true
}

// Decode copies the current document into v, which must be a pointer.
func (c *Cursor) Decode(v any) error {
	if c.DecodeErr != nil {
		return c.DecodeErr
	}
	if c.index < 0 || c.index >= len(c.docs) {
		return mongo.ErrNoDocuments
	}

	return convert(c.docs[c.index], v)
}

// Err returns nil: the stub does not fail while iterating.
func (c *Cursor) Err() error {
	return nil
}

// All copies the remaining documents into results, which must be a pointer to a slice, and closes the cursor.
func (c *Cursor) All(_ context.Context, results any) error {
	defer func() { c.closed = true }()

	if c.DecodeErr != nil {
		return c.DecodeErr
	}

	rv := reflect.ValueOf(results)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Slice {
		return &reflect.ValueError{Method: "mock.Cursor.All", Kind: rv.Kind()}
	}

	slice := reflect.MakeSlice(rv.Elem().Type(), 0, len(c.docs))
	for c.index+1 < len(c.docs) {
		c.index++

		elem := reflect.New(rv.Elem().Type().Elem())
		if err := convert(c.docs[c.index], elem.Interface()); err != nil {
			return err
		}
		slice = reflect.Append(slice, elem.Elem())
	}
	rv.Elem().Set(slice)

	return nil
}

// Close marks the cursor as closed and returns CloseErr.
func (c *Cursor) Close(context.Context) error {
	c.closed = true
	return c.CloseErr
}

// convert copies src into dst, a pointer, by way of BSON.
func convert(src, dst any) error {
	data, err := bson.Marshal(src)
	if err != nil {
		return err
	}

	return bson.Unmarshal(data, dst)
}

// IndexView is a stub of store.IndexView recording the indexes and the names it receives.
type IndexView struct {
	// Err is returned by every method.
	Err error

	mu      sync.Mutex
	Created []mongo.IndexModel
	Dropped []string
	AllDrop int
}

// CreateOne records the index model.
func (v *IndexView) CreateOne(_ context.Context, model mongo.IndexModel, _ ...options.Lister[options.CreateIndexesOptions]) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Created = append(v.Created, model)

	return "index", v.Err
}

// DropOne records the name of the index.
func (v *IndexView) DropOne(_ context.Context, name string, _ ...options.Lister[options.DropIndexesOptions]) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Dropped = append(v.Dropped, name)

	return v.Err
}

// DropAll counts the calls.
func (v *IndexView) DropAll(context.Context, ...options.Lister[options.DropIndexesOptions]) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.AllDrop++

	return v.Err
}
