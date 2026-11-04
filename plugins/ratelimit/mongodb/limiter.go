package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RateLimiter is a request.RateLimiter keeping its counters in the database, shared by the instances of the API. It
// counts the requests of each client in fixed windows of the period: a client can send up to twice the limit around the
// end of a window. The counters expire with their window, see Settings.Collection.
type RateLimiter struct {
	database   func() store.Database
	collection string
	name       string
	requests   int
	period     time.Duration
	now        func() time.Time
}

// NewRateLimiters returns the function creating the limiters of the groups of routes, to set as
// api.Server.Hooks.NewRateLimiter. The database is the one of the plugin, the counters are in the collection.
func NewRateLimiters(database func() store.Database, collection string) request.NewRateLimiterFunc {
	return func(name string, requests int, period time.Duration) request.RateLimiter {
		return &RateLimiter{database: database, collection: collection, name: name, requests: requests, period: period, now: time.Now}
	}
}

// Allow counts a request of the client and reports whether it is within the limit, and the delay to wait for
// otherwise, until the end of the window.
func (l *RateLimiter) Allow(ctx context.Context, client string) (ok bool, retryAfter time.Duration, err error) {
	now := l.now()
	window := now.Truncate(l.period)
	end := window.Add(l.period)

	filter := bson.M{"_id": fmt.Sprintf("%s/%s/%d", l.name, client, window.UnixMilli())}
	update := bson.M{"$inc": bson.M{"count": 1}, "$setOnInsert": bson.M{"expireAt": end}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter struct {
		Count int `bson:"count"`
	}

	collection := l.database().Collection(l.collection)
	err = collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if mongo.IsDuplicateKeyError(err) {
		// Two instances created the counter at the same time: the second update increments it.
		err = collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	}
	if err != nil {
		return false, 0, fmt.Errorf("unable to count the request: %w", err)
	}

	if counter.Count > l.requests {
		return false, end.Sub(now), nil
	}

	return true, 0, nil
}
