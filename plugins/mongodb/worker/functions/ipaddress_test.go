package functions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
	"github.com/cygy/ginamite/worker/tasks/async/location"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestSaveIPAddressDetails(t *testing.T) {
	ctx := context.Background()
	tokenID := bson.NewObjectID()
	locationID := bson.NewObjectID()

	details := &location.IPAddressDetails{
		Region: "IDF", RegionCode: "FR-IDF", City: "Paris", CountryCode: "FR", CountryName: "France",
		Source: "icloud-private-relay",
	}
	var asked []string
	lookup := func(ip string) (*location.IPAddressDetails, error) {
		asked = append(asked, ip)
		return details, nil
	}

	newDB := func() (*Functions, *mock.Database) {
		fns, db := withMainSession(t)
		db.Coll(model.AuthTokenCollection).One = mock.Result{Value: bson.M{"_id": tokenID}}

		return fns, db
	}

	t.Run("unknown location is looked up and saved", func(t *testing.T) {
		asked = nil
		fns, db := newDB()

		if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup); err != nil {
			t.Fatalf("error = %v", err)
		}

		saved := db.Coll(model.IPLocationCollection).LastCall().Args[0].(*model.IPLocation)
		if len(asked) != 1 || saved.IP != "1.2.3.4" || saved.Region != "IDF" || saved.City != "Paris" || saved.Country.Code != "FR" || saved.Country.Name != "France" ||
			saved.RegionCode != "FR-IDF" || saved.Source != "icloud-private-relay" || saved.UpdatedAt.IsZero() {
			t.Errorf("asked = %v, location = %+v", asked, saved)
		}

		update := db.Coll(model.AuthTokenCollection).LastCall()
		if update.Method != "UpdateByID" || update.Args[0] != tokenID {
			t.Errorf("token call = %+v", update)
		}
		if got := update.Args[1].(bson.M)["$set"].(bson.M)["location"].(model.IPLocation); got.City != "Paris" {
			t.Errorf("location = %+v", got)
		}
	})

	t.Run("fresh location is not looked up", func(t *testing.T) {
		asked = nil
		fns, db := newDB()
		db.Coll(model.IPLocationCollection).One = mock.Result{Value: bson.M{
			"_id": locationID, "IP": "1.2.3.4", "city": "Lyon", "source": "dbip", "updatedAt": time.Now().Add(-time.Hour),
		}}

		if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup); err != nil {
			t.Fatalf("error = %v", err)
		}

		if len(asked) != 0 {
			t.Errorf("asked = %v", asked)
		}
		for _, call := range db.Coll(model.IPLocationCollection).Calls() {
			if call.Method == "InsertOne" {
				t.Error("the location must not be saved again")
			}
		}
		if got := db.Coll(model.AuthTokenCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["location"].(model.IPLocation); got.City != "Lyon" {
			t.Errorf("location = %+v", got)
		}
	})

	for name, cached := range map[string]bson.M{
		"saved before the date was kept": {"_id": locationID, "IP": "1.2.3.4", "city": "Lyon"},
		"older than one day": {
			"_id": locationID, "IP": "1.2.3.4", "city": "Lyon", "source": "dbip",
			"updatedAt": time.Now().Add(-25 * time.Hour),
		},
	} {
		t.Run("stale location "+name+" is asked again and updated", func(t *testing.T) {
			asked = nil
			fns, db := newDB()
			db.Coll(model.IPLocationCollection).One = mock.Result{Value: cached}
			before := time.Now()

			if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup); err != nil {
				t.Fatalf("error = %v", err)
			}

			if len(asked) != 1 {
				t.Errorf("asked = %v, want the IP once", asked)
			}
			update := db.Coll(model.IPLocationCollection).LastCall()
			if update.Method != "UpdateByID" || update.Args[0] != locationID {
				t.Fatalf("location call = %+v, want an update", update)
			}
			set := update.Args[1].(bson.M)["$set"].(bson.M)
			if set["source"] != "icloud-private-relay" || set["city"] != "Paris" || set["regionCode"] != "FR-IDF" ||
				set["updatedAt"].(time.Time).Before(before) {
				t.Errorf("update = %+v", set)
			}
			if got := db.Coll(model.AuthTokenCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["location"].(model.IPLocation); got.City != "Paris" || got.Source != "icloud-private-relay" {
				t.Errorf("location of the token = %+v", got)
			}
		})
	}

	t.Run("stale location is kept when the API finds nothing", func(t *testing.T) {
		fns, db := newDB()
		db.Coll(model.IPLocationCollection).One = mock.Result{Value: bson.M{"_id": locationID, "IP": "1.2.3.4", "city": "Lyon"}}

		if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), func(string) (*location.IPAddressDetails, error) { return nil, nil }); err != nil {
			t.Fatalf("error = %v", err)
		}

		if got := methods(db, model.IPLocationCollection); len(got) != 1 {
			t.Errorf("location calls = %v, want only the lookup", got)
		}
		if got := db.Coll(model.AuthTokenCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["location"].(model.IPLocation); got.City != "Lyon" {
			t.Errorf("location of the token = %+v", got)
		}
	})

	t.Run("nothing to do", func(t *testing.T) {
		// The token does not exist anymore.
		fns, db := withMainSession(t)
		if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup); err != nil {
			t.Errorf("error = %v, a token deleted is not a failure", err)
		}
		if len(db.Coll(model.IPLocationCollection).Calls()) != 0 {
			t.Error("the location must not be looked up for an unknown token")
		}

		// The lookup of the location fails, or finds nothing.
		fns, db = newDB()
		if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), func(string) (*location.IPAddressDetails, error) { return nil, nil }); err != nil {
			t.Errorf("error = %v", err)
		}
		if got := methods(db, model.AuthTokenCollection); len(got) != 1 {
			t.Errorf("token calls = %v, want only the lookup", got)
		}

		// The API fails: the error is returned, to try again.
		fns, db = newDB()
		if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), func(string) (*location.IPAddressDetails, error) { return nil, errBoom }); !errors.Is(err, errBoom) {
			t.Errorf("error = %v, want the error of the API", err)
		}
		if got := methods(db, model.AuthTokenCollection); len(got) != 1 {
			t.Errorf("token calls = %v, want only the lookup", got)
		}
	})

	t.Run("a token deleted meanwhile is not a failure", func(t *testing.T) {
		fns, db := newDB()
		db.Coll(model.AuthTokenCollection).Errs = map[string]error{"UpdateByID": mongo.ErrNoDocuments}

		if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup); err != nil {
			t.Errorf("error = %v", err)
		}
	})

	failures := []struct {
		name    string
		prepare func(db *mock.Database)
	}{
		{"read of the token", func(db *mock.Database) { db.Coll(model.AuthTokenCollection).One = mock.Result{Error: errBoom} }},
		{"read of the location", func(db *mock.Database) {
			db.Coll(model.IPLocationCollection).Errs = map[string]error{"FindOne": errBoom}
		}},
		{"save of the location", func(db *mock.Database) {
			db.Coll(model.IPLocationCollection).Errs = map[string]error{"InsertOne": errBoom}
		}},
		{"update of the location", func(db *mock.Database) {
			db.Coll(model.IPLocationCollection).One = mock.Result{Value: bson.M{"_id": bson.NewObjectID(), "IP": "1.2.3.4"}}
			db.Coll(model.IPLocationCollection).Errs = map[string]error{"UpdateByID": errBoom}
		}},
		{"update of the token", func(db *mock.Database) {
			db.Coll(model.AuthTokenCollection).Errs = map[string]error{"UpdateByID": errBoom}
		}},
	}
	for _, tt := range failures {
		t.Run("failure of the "+tt.name, func(t *testing.T) {
			fns, db := newDB()
			tt.prepare(db)

			if err := fns.SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup); !errors.Is(err, errBoom) {
				t.Errorf("error = %v, want the error of the database", err)
			}
		})
	}
}
