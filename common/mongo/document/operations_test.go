package document

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/bson"

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

func (d *fakeDocument) IsSaved() bool                                          { return d.ID != "" }
func (d *fakeDocument) Save(context.Context, *mongo.Database) (string, error)  { return d.ID, nil }
func (d *fakeDocument) Update(context.Context, any, *mongo.Database) error     { return nil }
func (d *fakeDocument) Delete(context.Context, *mongo.Database) error          { return nil }
func (d *fakeDocument) GetByID(context.Context, string, *mongo.Database) error { return nil }
func (d *fakeDocument) GetID() string                                          { return d.ID }

func (d *fakeDocument) SetID(id string) error {
	if d.setIDErr != nil {
		return d.setIDErr
	}
	d.ID = id

	return nil
}

// offlineDB returns a database whose server can not be reached.
func offlineDB(t *testing.T) *mongo.Database {
	t.Helper()

	client, err := mongo.Connect(options.Client().
		SetHosts([]string{"127.0.0.1:1"}).
		SetServerSelectionTimeout(20 * time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	return client.Database("test")
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
