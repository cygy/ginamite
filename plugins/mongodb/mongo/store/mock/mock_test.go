package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	_ store.Database   = (*Database)(nil)
	_ store.Collection = (*Collection)(nil)
	_ store.Cursor     = (*Cursor)(nil)
	_ store.IndexView  = (*IndexView)(nil)
)

type item struct {
	Name string `bson:"name"`
}

func TestDatabase(t *testing.T) {
	db := NewDatabase("app")

	if db.Name() != "app" {
		t.Errorf("Name() = %q", db.Name())
	}
	if db.Collection("users") != store.Collection(db.Coll("users")) {
		t.Error("Collection() must return the same stub for a name")
	}
	if db.Collection("users") == db.Collection("posts") {
		t.Error("each name has its own collection")
	}

	// A zero Database is usable.
	var zero Database
	if zero.Coll("x") == nil {
		t.Error("a zero Database must create the collections")
	}
}

func TestDatabase_CreateCollection(t *testing.T) {
	db := NewDatabase("app")
	boom := errors.New("boom")

	opts := options.CreateCollection()
	if err := db.CreateCollection(context.Background(), "a", opts); err != nil {
		t.Fatal(err)
	}
	if got := db.CreateOptions("a"); len(got) != 1 || got[0] != options.Lister[options.CreateCollectionOptions](opts) {
		t.Errorf("CreateOptions() = %v", got)
	}
	db.CreateCollectionErr = boom
	if err := db.CreateCollection(context.Background(), "b"); !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}

	if got := db.Created(); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("Created() = %v", got)
	}
}

func TestCollection_Defaults(t *testing.T) {
	ctx := context.Background()
	c := NewDatabase("app").Coll("users")

	if r, err := c.InsertOne(ctx, item{}); err != nil || r == nil {
		t.Errorf("InsertOne = %v, %v", r, err)
	}
	for name, call := range map[string]func() (*mongo.UpdateResult, error){
		"UpdateByID": func() (*mongo.UpdateResult, error) { return c.UpdateByID(ctx, 1, nil) },
		"UpdateOne":  func() (*mongo.UpdateResult, error) { return c.UpdateOne(ctx, nil, nil) },
		"UpdateMany": func() (*mongo.UpdateResult, error) { return c.UpdateMany(ctx, nil, nil) },
	} {
		if r, err := call(); err != nil || r.MatchedCount != 1 || r.ModifiedCount != 1 {
			t.Errorf("%s = %+v, %v", name, r, err)
		}
	}
	for name, call := range map[string]func() (*mongo.DeleteResult, error){
		"DeleteOne":  func() (*mongo.DeleteResult, error) { return c.DeleteOne(ctx, nil) },
		"DeleteMany": func() (*mongo.DeleteResult, error) { return c.DeleteMany(ctx, nil) },
	} {
		if r, err := call(); err != nil || r.DeletedCount != 1 {
			t.Errorf("%s = %+v, %v", name, r, err)
		}
	}

	if err := c.FindOne(ctx, nil).Err(); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Errorf("FindOne().Err() = %v", err)
	}
	if n, err := c.CountDocuments(ctx, nil); n != 0 || err != nil {
		t.Errorf("CountDocuments = %d, %v", n, err)
	}
	if err := c.Drop(ctx); err != nil {
		t.Errorf("Drop = %v", err)
	}

	if got := len(c.Calls()); got != 9 {
		t.Errorf("recorded %d calls", got)
	}
	if c.LastCall().Method != "Drop" {
		t.Errorf("LastCall() = %+v", c.LastCall())
	}
	if (&Collection{}).LastCall().Method != "" {
		t.Error("LastCall() of a collection without call must be zero")
	}
}

func TestCollection_Configured(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("boom")

	c := &Collection{
		InsertResult: &mongo.InsertOneResult{InsertedID: 7},
		UpdateResult: &mongo.UpdateResult{MatchedCount: 4},
		DeleteResult: &mongo.DeleteResult{DeletedCount: 5},
		One:          Result{Value: item{Name: "one"}},
		Docs:         []any{item{Name: "a"}, item{Name: "b"}},
		Count:        9,
		CloseErr:     errors.New("close"),
	}

	if r, _ := c.InsertOne(ctx, nil); r.InsertedID != 7 {
		t.Errorf("InsertOne = %+v", r)
	}
	if r, _ := c.UpdateOne(ctx, nil, nil); r.MatchedCount != 4 {
		t.Errorf("UpdateOne = %+v", r)
	}
	if r, _ := c.DeleteOne(ctx, nil); r.DeletedCount != 5 {
		t.Errorf("DeleteOne = %+v", r)
	}
	if n, _ := c.CountDocuments(ctx, nil); n != 9 {
		t.Errorf("CountDocuments = %d", n)
	}

	var one item
	if err := c.FindOne(ctx, nil).Decode(&one); err != nil || one.Name != "one" {
		t.Errorf("FindOne = %+v, %v", one, err)
	}

	for name, query := range map[string]func() (store.Cursor, error){
		"Find":      func() (store.Cursor, error) { return c.Find(ctx, nil) },
		"Aggregate": func() (store.Cursor, error) { return c.Aggregate(ctx, nil) },
	} {
		cursor, err := query()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		var all []item
		if err := cursor.All(ctx, &all); err != nil || len(all) != 2 || all[1].Name != "b" {
			t.Errorf("%s: %+v, %v", name, all, err)
		}
		if err := cursor.Close(ctx); err == nil || err.Error() != "close" {
			t.Errorf("%s: Close err = %v", name, err)
		}
	}

	c.Err = boom
	if _, err := c.InsertOne(ctx, nil); !errors.Is(err, boom) {
		t.Errorf("InsertOne err = %v", err)
	}
	if _, err := c.UpdateByID(ctx, 1, nil); !errors.Is(err, boom) {
		t.Errorf("UpdateByID err = %v", err)
	}
	if _, err := c.DeleteMany(ctx, nil); !errors.Is(err, boom) {
		t.Errorf("DeleteMany err = %v", err)
	}
	if err := c.FindOne(ctx, nil).Err(); !errors.Is(err, boom) {
		t.Errorf("FindOne err = %v", err)
	}
	if cursor, err := c.Find(ctx, nil); !errors.Is(err, boom) || cursor != nil {
		t.Errorf("Find = %v, %v", cursor, err)
	}
	if cursor, err := c.Aggregate(ctx, nil); !errors.Is(err, boom) || cursor != nil {
		t.Errorf("Aggregate = %v, %v", cursor, err)
	}
	if _, err := c.CountDocuments(ctx, nil); !errors.Is(err, boom) {
		t.Errorf("CountDocuments err = %v", err)
	}
	if err := c.Drop(ctx); !errors.Is(err, boom) {
		t.Errorf("Drop err = %v", err)
	}
}

func TestCollection_RecordsArguments(t *testing.T) {
	ctx := context.Background()
	c := &Collection{}

	filter := bson.M{"a": 1}
	_, _ = c.Find(ctx, filter, options.Find().SetSkip(3).SetLimit(4).SetSort(bson.D{{Key: "a", Value: -1}}))
	_ = c.FindOne(ctx, filter, options.FindOne().SetSort(bson.D{{Key: "b", Value: 1}}))
	_, _ = c.UpdateOne(ctx, filter, bson.M{"$set": 1}, options.UpdateOne().SetUpsert(true))

	calls := c.Calls()
	if len(calls) != 3 {
		t.Fatalf("calls = %+v", calls)
	}

	find := FindOptions(calls[0].Opts)
	if *find.Skip != 3 || *find.Limit != 4 || find.Sort == nil {
		t.Errorf("find options = %+v", find)
	}
	if got := calls[0].Args[0].(bson.M); got["a"] != 1 {
		t.Errorf("filter = %v", got)
	}
	if FindOneOptions(calls[1].Opts).Sort == nil {
		t.Error("the FindOne sort must be recorded")
	}
	if !*UpdateOneOptions(calls[2].Opts).Upsert {
		t.Error("the upsert option must be recorded")
	}

	// The options of another type are ignored.
	if FindOptions([]any{"nope", options.FindOne()}).Skip != nil {
		t.Error("foreign options must be ignored")
	}
}

func TestIndexes(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("boom")
	c := &Collection{}

	view := c.Indexes()
	if view != store.IndexView(c.IndexView) || c.Indexes() != view {
		t.Error("Indexes() must always return the same view")
	}

	if _, err := view.CreateOne(ctx, mongo.IndexModel{}); err != nil {
		t.Fatal(err)
	}
	if err := view.DropOne(ctx, "idx"); err != nil {
		t.Fatal(err)
	}
	if err := view.DropAll(ctx); err != nil {
		t.Fatal(err)
	}
	if len(c.IndexView.Created) != 1 || c.IndexView.Dropped[0] != "idx" || c.IndexView.AllDrop != 1 {
		t.Errorf("view = %+v", c.IndexView)
	}

	c.IndexView.Err = boom
	if _, err := view.CreateOne(ctx, mongo.IndexModel{}); !errors.Is(err, boom) {
		t.Errorf("CreateOne err = %v", err)
	}
	if err := view.DropOne(ctx, "idx"); !errors.Is(err, boom) {
		t.Errorf("DropOne err = %v", err)
	}
	if err := view.DropAll(ctx); !errors.Is(err, boom) {
		t.Errorf("DropAll err = %v", err)
	}
}

func TestResult(t *testing.T) {
	boom := errors.New("boom")

	var got item
	if err := (Result{Value: item{Name: "x"}}).Decode(&got); err != nil || got.Name != "x" {
		t.Errorf("Decode = %+v, %v", got, err)
	}
	if err := (Result{}).Decode(&got); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Errorf("empty Decode err = %v", err)
	}
	if r := (Result{Error: boom}); !errors.Is(r.Err(), boom) || !errors.Is(r.Decode(&got), boom) {
		t.Error("the error must be returned by Err and Decode")
	}
	if err := (Result{Value: make(chan int)}).Decode(&got); err == nil {
		t.Error("an unencodable value must fail")
	}
}

func TestCursor(t *testing.T) {
	ctx := context.Background()

	t.Run("iterates", func(t *testing.T) {
		c := NewCursor(item{Name: "a"}, item{Name: "b"})

		var got item
		if err := c.Decode(&got); !errors.Is(err, mongo.ErrNoDocuments) {
			t.Errorf("Decode before Next err = %v", err)
		}

		var names []string
		for c.Next(ctx) {
			if err := c.Decode(&got); err != nil {
				t.Fatal(err)
			}
			names = append(names, got.Name)
		}
		if len(names) != 2 || names[0] != "a" || names[1] != "b" {
			t.Errorf("names = %v", names)
		}
		if c.Err() != nil {
			t.Errorf("Err() = %v", c.Err())
		}
	})

	t.Run("All", func(t *testing.T) {
		c := NewCursor(item{Name: "a"}, item{Name: "b"})
		c.Next(ctx)

		var all []item
		if err := c.All(ctx, &all); err != nil || len(all) != 1 || all[0].Name != "b" {
			t.Errorf("All = %+v, %v", all, err)
		}
		if !c.Closed() {
			t.Error("All must close the cursor")
		}
	})

	t.Run("All errors", func(t *testing.T) {
		boom := errors.New("boom")

		c := NewCursor(item{})
		c.DecodeErr = boom
		var all []item
		if err := c.All(ctx, &all); !errors.Is(err, boom) {
			t.Errorf("All err = %v", err)
		}
		if err := c.Decode(&all); !errors.Is(err, boom) {
			t.Errorf("Decode err = %v", err)
		}

		var notSlice item
		if err := NewCursor(item{}).All(ctx, &notSlice); err == nil {
			t.Error("a pointer to a non-slice must be rejected")
		}
		if err := NewCursor(item{}).All(ctx, all); err == nil {
			t.Error("a non-pointer must be rejected")
		}
		if err := NewCursor(make(chan int)).All(ctx, &all); err == nil {
			t.Error("an unencodable document must fail")
		}
	})

	t.Run("Close", func(t *testing.T) {
		boom := errors.New("boom")

		c := NewCursor()
		c.CloseErr = boom
		if err := c.Close(ctx); !errors.Is(err, boom) || !c.Closed() {
			t.Errorf("Close err = %v, closed = %v", err, c.Closed())
		}
	})
}

func TestCollection_ErrorsByMethod(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("boom")

	c := &Collection{Errs: map[string]error{"CountDocuments": boom}}

	if _, err := c.InsertOne(ctx, nil); err != nil {
		t.Errorf("InsertOne err = %v", err)
	}
	if _, err := c.CountDocuments(ctx, nil); !errors.Is(err, boom) {
		t.Errorf("CountDocuments err = %v", err)
	}

	// Err applies to the other methods.
	other := errors.New("other")
	c.Err = other
	if _, err := c.InsertOne(ctx, nil); !errors.Is(err, other) {
		t.Errorf("InsertOne err = %v", err)
	}
	if _, err := c.CountDocuments(ctx, nil); !errors.Is(err, boom) {
		t.Errorf("CountDocuments err = %v", err)
	}
}
