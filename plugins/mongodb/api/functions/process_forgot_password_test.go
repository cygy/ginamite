package functions

import (
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

func TestNewForgotPasswordProcessForUser(t *testing.T) {
	userID := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	id, err := NewForgotPasswordProcessForUser(c, userID.Hex(), "private-key")
	if err != nil || id == "" {
		t.Fatalf("got %q, %v", id, err)
	}

	methods := calls(db, model.ForgotPasswordProcessCollection)
	if len(methods) != 2 || methods[0] != "DeleteMany" || methods[1] != "InsertOne" {
		t.Fatalf("calls = %v", methods)
	}
	saved := db.Coll(model.ForgotPasswordProcessCollection).LastCall().Args[0].(*model.ForgotPasswordProcess)
	if saved.UserID != userID || saved.Code != "private-key" || saved.GetID() != id {
		t.Errorf("process = %+v", saved)
	}

	if _, err := NewForgotPasswordProcessForUser(c, "nope", "k"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	db.Coll(model.ForgotPasswordProcessCollection).Errs = map[string]error{"DeleteMany": errBoom}
	if _, err := NewForgotPasswordProcessForUser(c, userID.Hex(), "k"); err == nil {
		t.Error("a failure to delete the existing process must be reported")
	}

	db.Coll(model.ForgotPasswordProcessCollection).Errs = map[string]error{"InsertOne": errBoom}
	if _, err := NewForgotPasswordProcessForUser(c, userID.Hex(), "k"); err == nil {
		t.Error("a failure to save the process must be reported")
	}
}

func TestGetForgotPasswordProcessDetailsByID(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.ForgotPasswordProcessCollection).One = mock.Result{Value: bson.M{"_id": id, "userId": userID, "code": "key"}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	gotUser, key, err := GetForgotPasswordProcessDetailsByID(c, id.Hex())
	if err != nil || gotUser != userID.Hex() || key != "key" {
		t.Errorf("got %q, %q, %v", gotUser, key, err)
	}

	if _, _, err := GetForgotPasswordProcessDetailsByID(c, "nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}

	c, _ = plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
	if _, _, err := GetForgotPasswordProcessDetailsByID(c, id.Hex()); err == nil {
		t.Error("an unknown process must be reported")
	}
}

func TestWrongAccessToForgotPasswordProcess(t *testing.T) {
	id := bson.NewObjectID()

	newDB := func(tries int64) *mock.Database {
		db := mock.NewDatabase("app")
		db.Coll(model.ForgotPasswordProcessCollection).One = mock.Result{Value: bson.M{"_id": id, "tries": tries}}

		return db
	}

	t.Run("increments the tries", func(t *testing.T) {
		db := newDB(1)
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		WrongAccessToForgotPasswordProcess(c, id.Hex())
		checkUpdate(t, db, model.ForgotPasswordProcessCollection, id, bson.M{"$inc": bson.M{"tries": 1}})
	})

	t.Run("deletes the process after three wrong tries", func(t *testing.T) {
		db := newDB(2)
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		WrongAccessToForgotPasswordProcess(c, id.Hex())
		checkCall(t, db, model.ForgotPasswordProcessCollection, "DeleteOne")
	})

	t.Run("failures are only logged", func(t *testing.T) {
		for _, tries := range []int64{0, 2} {
			db := newDB(tries)
			db.Coll(model.ForgotPasswordProcessCollection).Errs = map[string]error{"UpdateByID": errBoom, "DeleteOne": errBoom}
			c, _ := plugintest.NewContext(db, "GET", "/", nil)

			WrongAccessToForgotPasswordProcess(c, id.Hex())
		}
	})

	t.Run("unknown process", func(t *testing.T) {
		db := mock.NewDatabase("app")
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		WrongAccessToForgotPasswordProcess(c, id.Hex())
		if got := calls(db, model.ForgotPasswordProcessCollection); len(got) != 1 {
			t.Errorf("calls = %v, want only the lookup", got)
		}
	})
}

func TestDeleteForgotPasswordProcess(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := DeleteForgotPasswordProcess(c, id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, model.ForgotPasswordProcessCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	if err := DeleteForgotPasswordProcess(c, "nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}
