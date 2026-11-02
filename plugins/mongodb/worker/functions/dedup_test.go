package functions

import (
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestDeduplicator_Handled(t *testing.T) {
	t.Run("unknown message", func(t *testing.T) {
		fns, db := withMainSession(t)

		if fns.Deduplicator().Handled("worker", "id-1") {
			t.Error("an unknown message must not be handled")
		}
		if got := methods(db, model.ProcessedMessageCollection); len(got) != 1 || got[0] != "CountDocuments" {
			t.Errorf("calls = %v", got)
		}
	})

	t.Run("known message", func(t *testing.T) {
		fns, db := withMainSession(t)
		db.Coll(model.ProcessedMessageCollection).Count = 1

		if !fns.Deduplicator().Handled("worker", "id-1") {
			t.Error("the message must be handled")
		}
	})

	t.Run("database failure", func(t *testing.T) {
		fns, db := withMainSession(t)
		db.Coll(model.ProcessedMessageCollection).Err = errBoom

		if fns.Deduplicator().Handled("worker", "id-1") {
			t.Error("a failure must let the message be handled again")
		}
	})
}

func TestDeduplicator_MarkHandled(t *testing.T) {
	t.Run("inserts the message with the group", func(t *testing.T) {
		fns, db := withMainSession(t)

		fns.Deduplicator().MarkHandled("worker", "id-1")

		calls := db.Coll(model.ProcessedMessageCollection).Calls()
		if len(calls) != 1 || calls[0].Method != "InsertOne" {
			t.Fatalf("calls = %+v", calls)
		}
	})

	t.Run("a message already remembered is not an error", func(t *testing.T) {
		fns, db := withMainSession(t)
		db.Coll(model.ProcessedMessageCollection).Err = mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000}}}

		fns.Deduplicator().MarkHandled("worker", "id-1")
	})

	t.Run("a failure is logged", func(t *testing.T) {
		fns, db := withMainSession(t)
		db.Coll(model.ProcessedMessageCollection).Err = errBoom

		fns.Deduplicator().MarkHandled("worker", "id-1")
	})
}

func TestDeduplicator_Key(t *testing.T) {
	d := deduplicator{functions: &Functions{}}
	if got := d.key("g", "i"); got != "g/i" {
		t.Errorf("key = %q", got)
	}
}
