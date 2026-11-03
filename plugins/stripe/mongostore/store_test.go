package mongostore

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/stripe"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	errTest = errors.New("test failure")
	now     = time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
)

// newStore returns a store on a mock database, with a fixed clock, and the mock collection.
func newStore(collection string) (*Store, *mock.Collection) {
	db := mock.NewDatabase("test")
	s := New(func() store.Database { return db }, collection)
	s.now = func() time.Time { return now }

	name := collection
	if name == "" {
		name = DefaultCollection
	}

	return s, db.Coll(name)
}

func TestNew(t *testing.T) {
	if s, _ := newStore(""); s.collection != DefaultCollection {
		t.Errorf("collection = %q, want %q", s.collection, DefaultCollection)
	}
	if s, _ := newStore("payments"); s.collection != "payments" {
		t.Errorf("collection = %q, want payments", s.collection)
	}
}

func TestStore_EnsureIndexes(t *testing.T) {
	s, coll := newStore("")
	coll.IndexView = &mock.IndexView{}

	if err := s.EnsureIndexes(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(coll.IndexView.Created) != 2 {
		t.Errorf("created %d indexes, want 2", len(coll.IndexView.Created))
	}

	coll.IndexView = &mock.IndexView{Err: errTest}
	if err := s.EnsureIndexes(context.Background()); !errors.Is(err, errTest) {
		t.Errorf("EnsureIndexes() = %v, want %v", err, errTest)
	}
}

func TestStore_Save(t *testing.T) {
	s, coll := newStore("")
	payment := &stripe.Payment{
		Reference: "ref1", CheckoutSessionID: "cs_1", UserID: "user1", Amount: 499, Currency: "eur",
		Status: stripe.StatusTodo, Metadata: map[string]string{"packID": "pack1"}, Data: "data", CreatedAt: now, UpdatedAt: now,
	}

	if err := s.Save(context.Background(), payment); err != nil {
		t.Fatal(err)
	}

	want := Payment{
		Reference: "ref1", CheckoutSessionID: "cs_1", UserID: "user1", Amount: 499, Currency: "eur",
		Status: stripe.StatusTodo, Metadata: map[string]string{"packID": "pack1"}, Data: "data", CreatedAt: now, UpdatedAt: now,
	}
	if call := coll.LastCall(); call.Method != "InsertOne" || !reflect.DeepEqual(call.Args[0], want) {
		t.Errorf("call = %+v", call)
	}

	coll.Err = errTest
	if err := s.Save(context.Background(), payment); !errors.Is(err, errTest) {
		t.Errorf("Save() = %v, want %v", err, errTest)
	}
}

func TestStore_Claim(t *testing.T) {
	s, coll := newStore("")

	claimed, err := s.Claim(context.Background(), "ref1")
	if err != nil || !claimed {
		t.Fatalf("Claim() = %v, %v, want true", claimed, err)
	}
	wantFilter := bson.M{"ref": "ref1", "status": bson.M{"$ne": stripe.StatusDone}}
	wantUpdate := bson.M{"$set": bson.M{"status": stripe.StatusDone, "updatedAt": now}}
	if call := coll.LastCall(); call.Method != "UpdateOne" || !reflect.DeepEqual(call.Args, []any{wantFilter, wantUpdate}) {
		t.Errorf("call = %+v", call)
	}

	coll.UpdateResult = &mongo.UpdateResult{}
	if claimed, err := s.Claim(context.Background(), "ref1"); err != nil || claimed {
		t.Errorf("Claim() = %v, %v, want false when no payment is updated", claimed, err)
	}

	coll.Err = errTest
	if _, err := s.Claim(context.Background(), "ref1"); !errors.Is(err, errTest) {
		t.Errorf("Claim() = %v, want %v", err, errTest)
	}
}

func TestStore_Release(t *testing.T) {
	s, coll := newStore("")

	if err := s.Release(context.Background(), "ref1"); err != nil {
		t.Fatal(err)
	}
	want := []any{bson.M{"ref": "ref1"}, bson.M{"$set": bson.M{"status": stripe.StatusTodo, "updatedAt": now}}}
	if call := coll.LastCall(); !reflect.DeepEqual(call.Args, want) {
		t.Errorf("call = %+v", call)
	}
}

func TestStore_Complete(t *testing.T) {
	s, coll := newStore("")

	if err := s.Complete(context.Background(), "ref1", "pi_1", "cus_1"); err != nil {
		t.Fatal(err)
	}
	want := []any{bson.M{"ref": "ref1"}, bson.M{"$set": bson.M{"pId": "pi_1", "cId": "cus_1", "updatedAt": now}}}
	if call := coll.LastCall(); !reflect.DeepEqual(call.Args, want) {
		t.Errorf("call = %+v", call)
	}
}

func TestStore_Get(t *testing.T) {
	s, coll := newStore("")

	if _, err := s.Get(context.Background(), "ref1"); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Errorf("Get() = %v, want %v", err, mongo.ErrNoDocuments)
	}

	coll.One = mock.Result{Value: Payment{Reference: "ref1", Amount: 499}}
	payment, err := s.Get(context.Background(), "ref1")
	if err != nil || payment.Reference != "ref1" || payment.Amount != 499 {
		t.Errorf("Get() = %+v, %v", payment, err)
	}
}
