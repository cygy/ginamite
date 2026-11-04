package mongodb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func newTestLimiter(t *testing.T, now time.Time) (*RateLimiter, *mock.Database) {
	t.Helper()

	db := mock.NewDatabase("test")
	limiter := NewRateLimiters(func() store.Database { return db }, DefaultCollection)("authentication", 3, time.Minute).(*RateLimiter)
	limiter.now = func() time.Time { return now }

	return limiter, db
}

func TestRateLimiter_Allow(t *testing.T) {
	now := time.Date(2026, 11, 2, 10, 0, 20, 0, time.UTC)

	tests := []struct {
		name      string
		count     int
		wantOK    bool
		wantRetry time.Duration
	}{
		{"first request", 1, true, 0},
		{"last request of the limit", 3, true, 0},
		{"over the limit: wait for the end of the window", 4, false, 40 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, db := newTestLimiter(t, now)
			db.Coll(DefaultCollection).One = mock.Result{Value: bson.M{"count": tt.count}}

			ok, retryAfter, err := limiter.Allow(context.Background(), "1.2.3.4")
			if err != nil || ok != tt.wantOK || retryAfter != tt.wantRetry {
				t.Fatalf("Allow() = %v, %v, %v", ok, retryAfter, err)
			}

			calls := db.Coll(DefaultCollection).Calls()
			if len(calls) != 1 || calls[0].Method != "FindOneAndUpdate" {
				t.Fatalf("calls = %+v", calls)
			}
			filter := calls[0].Args[0].(bson.M)
			if want := "authentication/1.2.3.4/" + "1793613600000"; filter["_id"] != want {
				t.Errorf("_id = %v, want %v", filter["_id"], want)
			}
			if opts := mock.FindOneAndUpdateOptions(calls[0].Opts); opts.Upsert == nil || !*opts.Upsert {
				t.Error("the counter must be created by the first request")
			}
		})
	}
}

func TestRateLimiter_Allow_Errors(t *testing.T) {
	boom := errors.New("boom")

	t.Run("database error", func(t *testing.T) {
		limiter, db := newTestLimiter(t, time.Now())
		db.Coll(DefaultCollection).Err = boom

		if _, _, err := limiter.Allow(context.Background(), "c"); !errors.Is(err, boom) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("duplicate key: the update is tried again", func(t *testing.T) {
		limiter, db := newTestLimiter(t, time.Now())
		db.Coll(DefaultCollection).Err = mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000}}}

		if _, _, err := limiter.Allow(context.Background(), "c"); err == nil {
			t.Error("a persistent duplicate key must fail")
		}
		if calls := db.Coll(DefaultCollection).Calls(); len(calls) != 2 {
			t.Errorf("calls = %d, want 2", len(calls))
		}
	})
}
