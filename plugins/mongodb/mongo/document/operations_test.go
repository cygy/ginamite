package document

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/log/logtest"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// These tests exercise the operations against a server that can not be reached:
// they check the error paths, the arguments checks and that no operation panics.
// The success paths need a MongoDB server.

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

// fakeDocument is a Document that keeps its ID.
type fakeDocument struct {
	ID       string
	setIDErr error
}

func (d *fakeDocument) IsSaved() bool                                         { return d.ID != "" }
func (d *fakeDocument) Save(context.Context, store.Database) (string, error)  { return d.ID, nil }
func (d *fakeDocument) Update(context.Context, any, store.Database) error     { return nil }
func (d *fakeDocument) Delete(context.Context, store.Database) error          { return nil }
func (d *fakeDocument) GetByID(context.Context, string, store.Database) error { return nil }
func (d *fakeDocument) GetID() string                                         { return d.ID }

func (d *fakeDocument) SetID(id string) error {
	if d.setIDErr != nil {
		return d.setIDErr
	}
	d.ID = id

	return nil
}

// offlineDB returns a database whose server can not be reached.
func offlineDB(t *testing.T) store.Database {
	t.Helper()

	client, err := mongo.Connect(options.Client().
		SetHosts([]string{"127.0.0.1:1"}).
		SetServerSelectionTimeout(20 * time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	return store.NewDatabase(client.Database("test"))
}

func TestSaveDocument(t *testing.T) {
	db := offlineDB(t)

	t.Run("generates the ID", func(t *testing.T) {
		doc := &fakeDocument{}

		id, err := SaveDocument(context.Background(), doc, "users", db)
		if err == nil {
			t.Error("expected the connection error")
		}
		if id == "" || id != doc.ID || bson.IsNilObjectID(id) {
			t.Errorf("id = %q, document ID = %q", id, doc.ID)
		}
	})

	t.Run("keeps the ID", func(t *testing.T) {
		doc := &fakeDocument{ID: bson.NewObjectID().Hex()}

		id, _ := SaveDocument(context.Background(), doc, "users", db)
		if id != doc.ID {
			t.Errorf("id = %q, want %q", id, doc.ID)
		}
	})

	t.Run("SetID error", func(t *testing.T) {
		boom := errors.New("boom")

		id, err := SaveDocument(context.Background(), &fakeDocument{setIDErr: boom}, "users", db)
		if !errors.Is(err, boom) || id != "" {
			t.Errorf("got %q, %v", id, err)
		}
	})
}

func TestOperations_Unreachable(t *testing.T) {
	db := offlineDB(t)
	ctx := context.Background()
	id := bson.NewObjectID()
	sel := bson.M{"a": 1}
	sort := []string{"-a"}
	fields := bson.M{"a": 1}

	var list []bson.M
	var one bson.M
	doc := &fakeDocument{}

	tests := []struct {
		name string
		call func() error
	}{
		{"UpdateDocument", func() error { return UpdateDocument(ctx, id, "c", bson.M{"$set": sel}, db) }},
		{"UpdateDocumentsBySelector", func() error { _, err := UpdateDocumentsBySelector(ctx, "c", sel, bson.M{"$set": sel}, db); return err }},
		{"UpsertDocumentBySelector", func() error { _, err := UpsertDocumentBySelector(ctx, "c", sel, bson.M{"$set": sel}, db); return err }},
		{"DeleteDocument", func() error { return DeleteDocument(ctx, id, "c", db) }},
		{"DeleteDocumentsBySelector", func() error { _, err := DeleteDocumentsBySelector(ctx, "c", sel, db); return err }},
		{"GetDocumentByID", func() error { return GetDocumentByID(ctx, &fakeDocument{}, "c", id.Hex(), db) }},
		{"GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit", func() error {
			return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, &list, "c", sel, fields, sort, 1, 2, db)
		}},
		{"GetDocumentsBySelectorAndSortWithOffsetAndLimit", func() error {
			return GetDocumentsBySelectorAndSortWithOffsetAndLimit(ctx, &list, "c", sel, sort, 1, 2, db)
		}},
		{"GetDocumentsBySelectorAndSort", func() error { return GetDocumentsBySelectorAndSort(ctx, &list, "c", sel, sort, db) }},
		{"GetDocumentsBySelector", func() error { return GetDocumentsBySelector(ctx, &list, "c", sel, db) }},
		{"GetDocumentsBySelectorWithFields", func() error { return GetDocumentsBySelectorWithFields(ctx, &list, "c", sel, fields, db) }},
		{"GetDocumentsBySelectorWithOffsetAndLimit", func() error {
			return GetDocumentsBySelectorWithOffsetAndLimit(ctx, &list, "c", sel, 1, 2, db)
		}},
		{"GetDocumentsBySort", func() error { return GetDocumentsBySort(ctx, &list, "c", sort, db) }},
		{"GetDocumentsBySortWithOffsetAndLimit", func() error { return GetDocumentsBySortWithOffsetAndLimit(ctx, &list, "c", sort, 1, 2, db) }},
		{"GetDocuments", func() error { return GetDocuments(ctx, &list, "c", db) }},
		{"GetDocumentsWithOffsetAndLimit", func() error { return GetDocumentsWithOffsetAndLimit(ctx, &list, "c", 1, 2, db) }},
		{"GetDocumentBySelectorAndSort", func() error { return GetDocumentBySelectorAndSort(ctx, doc, "c", sel, sort, db) }},
		{"GetDocumentBySelector", func() error { return GetDocumentBySelector(ctx, doc, "c", sel, db) }},
		{"GetDocumentBySort", func() error { return GetDocumentBySort(ctx, doc, "c", sort, db) }},
		{"GetDocument", func() error { return GetDocument(ctx, doc, "c", db) }},
		{"CountBySelector", func() error { _, err := CountBySelector(ctx, "c", sel, db); return err }},
		{"Count", func() error { _, err := Count(ctx, "c", db); return err }},
		{"PipeAll", func() error { return PipeAll(ctx, &list, "c", []bson.M{{"$match": sel}}, db) }},
		{"PipeOne", func() error { return PipeOne(ctx, &one, "c", []bson.M{{"$match": sel}}, db) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatal("expected the connection error")
			}
			if commonErrors.IsNotFound(err) || errors.Is(err, mongo.ErrNoDocuments) {
				t.Errorf("a connection failure must not be reported as not found: %v", err)
			}
		})
	}
}

func TestGetDocumentByID_InvalidID(t *testing.T) {
	err := GetDocumentByID(context.Background(), &fakeDocument{}, "c", "nope", offlineDB(t))
	if err == nil {
		t.Error("an invalid ID must be rejected")
	}
}

func TestGetDocumentsOptions(t *testing.T) {
	// The negative offsets and limits are ignored, the others are accepted by the driver.
	db := offlineDB(t)

	var list []bson.M
	for _, bounds := range [][2]int64{{-1, -1}, {0, 0}, {5, -1}, {-1, 5}} {
		if err := GetDocumentsWithOffsetAndLimit(context.Background(), &list, "c", bounds[0], bounds[1], db); err == nil {
			t.Errorf("offset %d, limit %d: expected the connection error", bounds[0], bounds[1])
		}
	}
}

// The tests below use a stub database: they check the arguments sent to the driver and the results
// of the operations, but not the evaluation of the selectors, which needs a MongoDB server.

func TestSaveDocument_Insert(t *testing.T) {
	db := mock.NewDatabase("app")
	doc := &fakeDocument{}

	id, err := SaveDocument(context.Background(), doc, "users", db)
	if err != nil || id != doc.ID || id == "" {
		t.Fatalf("got %q, %v", id, err)
	}

	call := db.Coll("users").LastCall()
	if call.Method != "InsertOne" || call.Args[0] != doc {
		t.Errorf("call = %+v", call)
	}
}

func TestSaveDocument_InsertError(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll("users").Err = errors.New("boom")

	if _, err := SaveDocument(context.Background(), &fakeDocument{}, "users", db); err == nil {
		t.Error("expected the insert error")
	}
}

func TestUpdateDocument(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()
	update := bson.M{"$set": bson.M{"a": 1}}

	t.Run("updates", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := UpdateDocument(ctx, id, "users", update, db); err != nil {
			t.Fatal(err)
		}

		call := db.Coll("users").LastCall()
		if call.Method != "UpdateByID" || call.Args[0] != id {
			t.Errorf("call = %+v", call)
		}
	})

	t.Run("no match", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll("users").UpdateResult = &mongo.UpdateResult{}

		if err := UpdateDocument(ctx, id, "users", update, db); !errors.Is(err, mongo.ErrNoDocuments) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll("users").Err = errors.New("boom")

		if err := UpdateDocument(ctx, id, "users", update, db); err == nil {
			t.Error("expected the update error")
		}
	})
}

func TestUpdateDocumentsBySelector(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll("users").UpdateResult = &mongo.UpdateResult{MatchedCount: 3}

	n, err := UpdateDocumentsBySelector(ctx, "users", bson.M{"a": 1}, bson.M{"$set": 1}, db)
	if err != nil || n != 3 {
		t.Errorf("got %d, %v", n, err)
	}
	if got := db.Coll("users").LastCall().Method; got != "UpdateMany" {
		t.Errorf("method = %s", got)
	}

	db.Coll("users").Err = errors.New("boom")
	if n, err := UpdateDocumentsBySelector(ctx, "users", nil, nil, db); err == nil || n != 0 {
		t.Errorf("got %d, %v", n, err)
	}
}

func TestUpsertDocumentBySelector(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll("users").UpdateResult = &mongo.UpdateResult{MatchedCount: 1}

	n, err := UpsertDocumentBySelector(ctx, "users", bson.M{"a": 1}, bson.M{"$set": 1}, db)
	if err != nil || n != 1 {
		t.Errorf("got %d, %v", n, err)
	}

	call := db.Coll("users").LastCall()
	if call.Method != "UpdateOne" || !*mock.UpdateOneOptions(call.Opts).Upsert {
		t.Errorf("call = %+v, the upsert option is required", call)
	}

	db.Coll("users").Err = errors.New("boom")
	if n, err := UpsertDocumentBySelector(ctx, "users", nil, nil, db); err == nil || n != 0 {
		t.Errorf("got %d, %v", n, err)
	}
}

func TestDeleteDocument(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()

	t.Run("deletes", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := DeleteDocument(ctx, id, "users", db); err != nil {
			t.Fatal(err)
		}
		if call := db.Coll("users").LastCall(); call.Method != "DeleteOne" {
			t.Errorf("call = %+v", call)
		}
	})

	t.Run("no match", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll("users").DeleteResult = &mongo.DeleteResult{}

		if err := DeleteDocument(ctx, id, "users", db); !errors.Is(err, mongo.ErrNoDocuments) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestDeleteDocumentsBySelector(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll("users").DeleteResult = &mongo.DeleteResult{DeletedCount: 4}

	n, err := DeleteDocumentsBySelector(ctx, "users", bson.M{"a": 1}, db)
	if err != nil || n != 4 {
		t.Errorf("got %d, %v", n, err)
	}

	db.Coll("users").Err = errors.New("boom")
	if n, err := DeleteDocumentsBySelector(ctx, "users", nil, db); err == nil || n != 0 {
		t.Errorf("got %d, %v", n, err)
	}
}

func TestGetDocumentByID_Stub(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()

	tests := []struct {
		name         string
		one          store.SingleResult
		err          error
		wantNotFound bool
		wantID       string
	}{
		{"found", mock.Result{Value: bson.M{"id": "abc"}}, nil, false, "abc"},
		{"no document", nil, nil, true, ""},
		{"unsaved document", mock.Result{Value: bson.M{"other": 1}}, nil, true, ""},
		{"query error", nil, errors.New("boom"), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll("users").One = tt.one
			db.Coll("users").Err = tt.err

			doc := &fakeDocument{}
			err := GetDocumentByID(ctx, doc, "users", id.Hex(), db)

			switch {
			case tt.wantNotFound && !commonErrors.IsNotFound(err):
				t.Errorf("err = %v, want not found", err)
			case tt.err != nil && (err == nil || commonErrors.IsNotFound(err)):
				t.Errorf("err = %v, want the query error", err)
			case !tt.wantNotFound && tt.err == nil && err != nil:
				t.Errorf("err = %v", err)
			}
			if doc.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", doc.ID, tt.wantID)
			}
		})
	}
}

func TestGetDocuments_Stub(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll("users").Docs = []any{bson.M{"n": 1}, bson.M{"n": 2}}

	var list []bson.M
	err := GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, &list, "users", bson.M{"a": 1}, bson.M{"n": 1}, []string{"-n", "a"}, 2, 5, db)
	if err != nil || len(list) != 2 {
		t.Fatalf("got %v, %v", list, err)
	}

	opts := mock.FindOptions(db.Coll("users").LastCall().Opts)
	if *opts.Skip != 2 || *opts.Limit != 5 || opts.Projection == nil {
		t.Errorf("options = %+v", opts)
	}
	if want := (bson.D{{Key: "n", Value: -1}, {Key: "a", Value: 1}}); !reflect.DeepEqual(opts.Sort, want) {
		t.Errorf("sort = %v, want %v", opts.Sort, want)
	}
}

func TestGetDocuments_StubDefaults(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")

	var list []bson.M
	if err := GetDocuments(ctx, &list, "users", db); err != nil {
		t.Fatal(err)
	}

	// A negative offset or limit, and empty sort and fields, must not be sent.
	opts := mock.FindOptions(db.Coll("users").LastCall().Opts)
	if opts.Skip != nil || opts.Limit != nil || opts.Sort != nil || opts.Projection != nil {
		t.Errorf("options = %+v", opts)
	}
}

func TestGetDocuments_StubErrors(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("boom")

	db := mock.NewDatabase("app")
	db.Coll("users").Err = boom

	var list []bson.M
	if err := GetDocuments(ctx, &list, "users", db); !errors.Is(err, boom) {
		t.Errorf("find err = %v", err)
	}

	// The documents can not be decoded into a map of the wrong type.
	db = mock.NewDatabase("app")
	db.Coll("users").Docs = []any{bson.M{"n": 1}}

	var wrong []int
	if err := GetDocuments(ctx, &wrong, "users", db); err == nil {
		t.Error("expected the decoding error")
	}
}

func TestGetDocumentBySelectorAndSort_Stub(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll("users").One = mock.Result{Value: bson.M{"id": "xyz"}}

	doc := &fakeDocument{}
	if err := GetDocumentBySelectorAndSort(ctx, doc, "users", bson.M{"a": 1}, []string{"-a"}, db); err != nil || doc.ID != "xyz" {
		t.Fatalf("got %+v, %v", doc, err)
	}

	call := db.Coll("users").LastCall()
	if call.Method != "FindOne" || mock.FindOneOptions(call.Opts).Sort == nil {
		t.Errorf("call = %+v, the sort is required", call)
	}

	if err := GetDocument(ctx, doc, "users", db); err != nil {
		t.Fatal(err)
	}
	if mock.FindOneOptions(db.Coll("users").LastCall().Opts).Sort != nil {
		t.Error("no sort must be sent for an empty sort")
	}

	db.Coll("users").One = nil
	if err := GetDocumentBySelector(ctx, doc, "users", nil, db); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Errorf("err = %v", err)
	}
}

func TestCount_Stub(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll("users").Count = 12

	if n, err := CountBySelector(ctx, "users", bson.M{"a": 1}, db); err != nil || n != 12 {
		t.Errorf("got %d, %v", n, err)
	}
	if n, err := Count(ctx, "users", db); err != nil || n != 12 {
		t.Errorf("got %d, %v", n, err)
	}

	db.Coll("users").Err = errors.New("boom")
	if _, err := Count(ctx, "users", db); err == nil {
		t.Error("expected the count error")
	}
}

func TestPipeAll_Stub(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll("users").Docs = []any{bson.M{"n": 1}}

	var list []bson.M
	if err := PipeAll(ctx, &list, "users", []bson.M{}, db); err != nil || len(list) != 1 {
		t.Errorf("got %v, %v", list, err)
	}

	db.Coll("users").Err = errors.New("boom")
	if err := PipeAll(ctx, &list, "users", []bson.M{}, db); err == nil {
		t.Error("expected the aggregate error")
	}
}

func TestPipeOne_Stub(t *testing.T) {
	ctx := context.Background()

	t.Run("first result", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll("users").Docs = []any{bson.M{"n": 1}, bson.M{"n": 2}}

		var one bson.M
		if err := PipeOne(ctx, &one, "users", []bson.M{}, db); err != nil || one["n"] != int32(1) {
			t.Errorf("got %v, %v", one, err)
		}
	})

	t.Run("no result", func(t *testing.T) {
		var one bson.M
		if err := PipeOne(ctx, &one, "users", []bson.M{}, mock.NewDatabase("app")); !errors.Is(err, mongo.ErrNoDocuments) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("close error is only logged", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll("users").Docs = []any{bson.M{"n": 1}}
		db.Coll("users").CloseErr = errors.New("boom")

		var one bson.M
		if err := PipeOne(ctx, &one, "users", []bson.M{}, db); err != nil {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("aggregate error", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll("users").Err = errors.New("boom")

		var one bson.M
		if err := PipeOne(ctx, &one, "users", []bson.M{}, db); err == nil {
			t.Error("expected the aggregate error")
		}
	})
}

func TestNotFound(t *testing.T) {
	err := notFound(mongo.ErrNoDocuments)
	if !commonErrors.IsNotFound(err) || !errors.Is(err, mongo.ErrNoDocuments) {
		t.Errorf("notFound(ErrNoDocuments) = %v, want an error matching ErrNotFound and ErrNoDocuments", err)
	}

	if again := notFound(err); !errors.Is(again, err) || again.Error() != err.Error() {
		t.Errorf("notFound() must not wrap twice, got %v", again)
	}

	other := errors.New("boom")
	if got := notFound(other); !errors.Is(got, other) || got.Error() != other.Error() {
		t.Errorf("notFound(other) = %v, want it unchanged", got)
	}

	if got := notFound(nil); got != nil {
		t.Errorf("notFound(nil) = %v, want nil", got)
	}
}

func TestGetDocumentByID_NotFoundNotLogged(t *testing.T) {
	hook, restore := logtest.Replace(&log.Main, slog.LevelDebug)
	t.Cleanup(restore)

	// A missing document is an expected result, not a database error.
	err := GetDocumentByID(context.Background(), &fakeDocument{}, "users", bson.NewObjectID().Hex(), mock.NewDatabase("app"))
	if !commonErrors.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
	if entries := hook.AllEntries(); len(entries) != 0 {
		t.Errorf("logs = %v, want none", entries)
	}

	// A query error is logged.
	db := mock.NewDatabase("app")
	db.Coll("users").Err = errors.New("boom")
	if err := GetDocumentByID(context.Background(), &fakeDocument{}, "users", bson.NewObjectID().Hex(), db); err == nil {
		t.Fatal("the query error must be returned")
	}
	if len(hook.AllEntries()) != 1 {
		t.Errorf("logs = %v, want the database error", hook.AllEntries())
	}
}
