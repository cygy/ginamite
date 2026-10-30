package functions

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

func TestSaveIPAddressDetails(t *testing.T) {
	ctx := context.Background()
	tokenID := bson.NewObjectID()
	locationID := bson.NewObjectID()

	details := &location.IPAddressDetails{Region: "IDF", City: "Paris", CountryCode: "FR", CountryName: "France"}
	var asked []string
	lookup := func(ip string) *location.IPAddressDetails {
		asked = append(asked, ip)
		return details
	}

	newDB := func() *mock.Database {
		db := withMainSession(t)
		db.Coll(model.AuthTokenCollection).One = mock.Result{Value: bson.M{"_id": tokenID}}

		return db
	}

	t.Run("unknown location is looked up and saved", func(t *testing.T) {
		asked = nil
		db := newDB()

		SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup)

		saved := db.Coll(model.IPLocationCollection).LastCall().Args[0].(*model.IPLocation)
		if len(asked) != 1 || saved.IP != "1.2.3.4" || saved.Region != "IDF" || saved.City != "Paris" || saved.Country.Code != "FR" || saved.Country.Name != "France" {
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

	t.Run("known location is not looked up", func(t *testing.T) {
		asked = nil
		db := newDB()
		db.Coll(model.IPLocationCollection).One = mock.Result{Value: bson.M{"_id": locationID, "IP": "1.2.3.4", "city": "Lyon"}}

		SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup)

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

	t.Run("nothing to do", func(t *testing.T) {
		// The token does not exist anymore.
		db := withMainSession(t)
		SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup)
		if len(db.Coll(model.IPLocationCollection).Calls()) != 0 {
			t.Error("the location must not be looked up for an unknown token")
		}

		// The lookup of the location fails.
		db = newDB()
		SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), func(string) *location.IPAddressDetails { return nil })
		if got := methods(db, model.AuthTokenCollection); len(got) != 1 {
			t.Errorf("token calls = %v, want only the lookup", got)
		}

		// A failing query for the location.
		db = newDB()
		db.Coll(model.IPLocationCollection).Errs = map[string]error{"FindOne": errBoom}
		SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup)
		if got := methods(db, model.AuthTokenCollection); len(got) != 1 {
			t.Errorf("token calls = %v, want only the lookup", got)
		}
	})

	t.Run("failures to save are only logged", func(t *testing.T) {
		db := newDB()
		db.Coll(model.IPLocationCollection).Errs = map[string]error{"InsertOne": errBoom}
		db.Coll(model.AuthTokenCollection).Errs = map[string]error{"UpdateByID": errBoom}

		SaveIPAddressDetails(ctx, "1.2.3.4", tokenID.Hex(), lookup)
	})
}
