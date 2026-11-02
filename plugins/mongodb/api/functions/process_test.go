package functions

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	"github.com/gin-gonic/gin"
)

func TestCreateNewProcesses(t *testing.T) {
	userID := bson.NewObjectID()

	tests := []struct {
		name     string
		create   func(c *gin.Context, userID string) (string, string, error)
		wantType string
	}{
		{"delete account", CreateNewDeleteAccountProcess, authentication.ProcessDeleteAccount},
		{"disable account", CreateNewDisableAccountProcess, authentication.ProcessDisableAccount},
		{"verify email address", CreateNewVerifyEmailAddressProcess, authentication.ProcessVerifyEmailAddress},
		{"update password", func(c *gin.Context, userID string) (string, string, error) {
			return CreateNewUpdatePasswordProcess(c, userID, "hash")
		}, authentication.ProcessUpdatePassword},
		{"update email address", func(c *gin.Context, userID string) (string, string, error) {
			return CreateNewUpdateEmailAddressProcess(c, userID, "new@example.com")
		}, authentication.ProcessUpdateEmailAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			c, _ := plugintest.NewContext(db, "GET", "/", nil)

			id, code, err := tt.create(c, userID.Hex())
			if err != nil || id == "" || len(code) != 64 {
				t.Fatalf("got %q, %q, %v", id, code, err)
			}

			// The existing process is deleted, then the new one is saved.
			methods := calls(db, model.UpdatePropertyProcessCollection)
			if len(methods) != 2 || methods[0] != "DeleteMany" || methods[1] != "InsertOne" {
				t.Fatalf("calls = %v", methods)
			}
			saved := db.Coll(model.UpdatePropertyProcessCollection).LastCall().Args[0].(*model.UpdatePropertyProcess)
			if saved.Type != tt.wantType || saved.UserID != userID || saved.Code != code || saved.GetID() != id {
				t.Errorf("process = %+v", saved)
			}

			// Errors.
			if _, _, err := tt.create(c, "nope"); err == nil {
				t.Error("an invalid user ID must be rejected")
			}

			db.Coll(model.UpdatePropertyProcessCollection).Errs = map[string]error{"DeleteMany": errBoom}
			if _, _, err := tt.create(c, userID.Hex()); err == nil {
				t.Error("a failure to delete the existing process must be reported")
			}

			db.Coll(model.UpdatePropertyProcessCollection).Errs = map[string]error{"InsertOne": errBoom}
			if _, _, err := tt.create(c, userID.Hex()); err == nil {
				t.Error("a failure to save the process must be reported")
			}
		})
	}
}

func TestCreateProcess_NothingToDelete(t *testing.T) {
	// A "not found" error from the deletion is not a failure.
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	deleteNothing := func(context.Context, string, store.Database) error { return commonErrors.NotFound() }
	id, _, err := createProcess(c, bson.NewObjectID().Hex(), deleteNothing, func() (*model.UpdatePropertyProcess, error) {
		return model.NewDeleteAccountProcess(bson.NewObjectID().Hex())
	})
	if err != nil || id == "" {
		t.Errorf("got %q, %v", id, err)
	}
}

func TestVerifyProcess(t *testing.T) {
	processID := bson.NewObjectID()
	userID := bson.NewObjectID()

	newDB := func(tries uint) *mock.Database {
		db := mock.NewDatabase("app")
		db.Coll(model.UpdatePropertyProcessCollection).One = mock.Result{Value: bson.M{"_id": processID, "userId": userID, "code": "right", "value": "stored", "tries": int64(tries)}}

		return db
	}

	t.Run("right key", func(t *testing.T) {
		c, _ := plugintest.NewContext(newDB(0), "GET", "/", nil)
		gotUser, value, err := VerifyProcess(c, processID.Hex(), "right")
		if err != nil || gotUser != userID.Hex() || value != "stored" {
			t.Errorf("got %q, %q, %v", gotUser, value, err)
		}
	})

	t.Run("wrong key increments the tries", func(t *testing.T) {
		db := newDB(0)
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		if _, _, err := VerifyProcess(c, processID.Hex(), "wrong"); !commonErrors.IsNotFound(err) {
			t.Errorf("err = %v", err)
		}
		checkUpdate(t, db, model.UpdatePropertyProcessCollection, processID, bson.M{"$inc": bson.M{"tries": 1}})
	})

	t.Run("third wrong key deletes the process", func(t *testing.T) {
		db := newDB(2)
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		if _, _, err := VerifyProcess(c, processID.Hex(), "wrong"); !commonErrors.IsNotFound(err) {
			t.Errorf("err = %v", err)
		}
		checkCall(t, db, model.UpdatePropertyProcessCollection, "DeleteOne")
	})

	t.Run("failures to count the try are only logged", func(t *testing.T) {
		for _, tries := range []uint{0, 2} {
			db := newDB(tries)
			db.Coll(model.UpdatePropertyProcessCollection).Errs = map[string]error{"UpdateByID": errBoom, "DeleteOne": errBoom}
			c, _ := plugintest.NewContext(db, "GET", "/", nil)

			if _, _, err := VerifyProcess(c, processID.Hex(), "wrong"); !commonErrors.IsNotFound(err) {
				t.Errorf("tries %d: err = %v", tries, err)
			}
		}
	})

	t.Run("unknown process", func(t *testing.T) {
		c, _ := plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
		if _, _, err := VerifyProcess(c, processID.Hex(), "right"); !commonErrors.IsNotFound(err) {
			t.Errorf("err = %v", err)
		}
		if _, _, err := VerifyProcess(c, "nope", "right"); err == nil {
			t.Error("an invalid process ID must be rejected")
		}
	})
}

func TestDeleteProcessByID(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := DeleteProcessByID(c, id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, model.UpdatePropertyProcessCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	if err := DeleteProcessByID(c, "nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}
