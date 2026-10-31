package functions

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

func userDocument(id bson.ObjectID) bson.M {
	return bson.M{
		"_id": id, "username": "john", "privateKey": "key",
		"privateInfos":          bson.M{"email": "j@example.com", "isEmailValid": true},
		"settings":              bson.M{"locale": "fr"},
		"identificationMethods": bson.M{"password": "hash"},
	}
}

func TestGetUserDetailsByIdentifier(t *testing.T) {
	id := bson.NewObjectID()

	tests := []struct {
		identifier string
		filter     bson.M
	}{
		{"j@example.com", bson.M{"privateInfos.email": "j@example.com"}},
		{"john", bson.M{"username": "john"}},
	}
	for _, tt := range tests {
		t.Run(tt.identifier, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
			c, _ := plugintest.NewContext(db, "GET", "/", nil)

			user, err := GetUserDetailsByIdentifier(c, tt.identifier)
			if err != nil || user.ID != id.Hex() || user.Username != "john" || user.EmailAddress != "j@example.com" {
				t.Errorf("got %+v, %v", user, err)
			}
			checkCall(t, db, model.UserCollection, "FindOne", tt.filter)
		})
	}

	c, _ := plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
	if _, err := GetUserDetailsByIdentifier(c, "nobody"); err == nil {
		t.Error("an unknown user must be reported")
	}
}

func TestGetUserDetailsByID(t *testing.T) {
	id := bson.NewObjectID()

	t.Run("from the database", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		user, err := GetUserDetailsByID(c, id.Hex())
		if err != nil || user.Username != "john" {
			t.Errorf("got %+v, %v", user, err)
		}
	})

	t.Run("from the cache of the authenticated user", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
		c, _ := plugintest.SignedInContext(t, db, model.User{ID: id})

		for range 2 {
			if user, err := GetUserDetailsByID(c, id.Hex()); err != nil || user.Username != "john" {
				t.Fatalf("got %+v, %v", user, err)
			}
		}
		if n := len(db.Coll(model.UserCollection).Calls()); n != 1 {
			t.Errorf("%d queries, want 1: the user must be cached", n)
		}
	})

	t.Run("another user than the authenticated one", func(t *testing.T) {
		other := bson.NewObjectID()
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": other, "username": "jane"}}
		c, _ := plugintest.SignedInContext(t, db, model.User{ID: id})

		if user, err := GetUserDetailsByID(c, other.Hex()); err != nil || user.Username != "jane" {
			t.Errorf("got %+v, %v", user, err)
		}
	})

	t.Run("errors", func(t *testing.T) {
		c, _ := plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
		if _, err := GetUserDetailsByID(c, "nope"); err == nil {
			t.Error("an invalid ID must be rejected")
		}
		if _, err := GetUserDetailsByID(c, id.Hex()); err == nil {
			t.Error("an unknown user must be reported")
		}
	})
}

func TestGetUserDetailsBySocialNetworkID(t *testing.T) {
	id := bson.NewObjectID()

	tests := []struct {
		name   string
		call   func(c *gin.Context) (string, error)
		filter bson.M
	}{
		{"facebook", func(c *gin.Context) (string, error) {
			user, err := GetUserDetailsByFacebookUserID(c, "10157123")
			return user.Username, err
		}, bson.M{"identificationMethods.facebook.userId": "10157123"}},
		{"google", func(c *gin.Context) (string, error) {
			user, err := GetUserDetailsByGoogleUserID(c, "11800")
			return user.Username, err
		}, bson.M{"identificationMethods.google.userId": "11800"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
			c, _ := plugintest.NewContext(db, "GET", "/", nil)

			if username, err := tt.call(c); err != nil || username != "john" {
				t.Errorf("got %q, %v", username, err)
			}
			checkCall(t, db, model.UserCollection, "FindOne", tt.filter)

			db.Coll(model.UserCollection).Err = errBoom
			if _, err := tt.call(c); !errors.Is(err, errBoom) {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestGetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	user, password, err := GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(c, "john")
	if err != nil || user.ID != id.Hex() || password != "hash" {
		t.Errorf("got %+v, %q, %v", user, password, err)
	}

	c, _ = plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
	if _, _, err := GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(c, "nobody"); !commonErrors.IsNotFound(err) {
		t.Errorf("err = %v", err)
	}
}

func TestDeleteUserByID(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := DeleteUserByID(c, id.Hex()); err != nil {
		t.Fatal(err)
	}

	// The deletion is recorded before the user is deleted.
	deleted := db.Coll(model.DeletedUserCollection).LastCall().Args[0].(*model.DeletedUser)
	if deleted.UserID != id {
		t.Errorf("deleted user = %+v", deleted)
	}
	checkCall(t, db, model.UserCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	if err := DeleteUserByID(c, "nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}

	db = mock.NewDatabase("app")
	db.Coll(model.DeletedUserCollection).Err = errBoom
	c, _ = plugintest.NewContext(db, "GET", "/", nil)
	if err := DeleteUserByID(c, id.Hex()); err == nil {
		t.Error("a failure to record the deletion must be reported")
	}
	if got := calls(db, model.UserCollection); len(got) != 0 {
		t.Errorf("the user must not be deleted: %v", got)
	}
}

func TestDisableUserByID(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	user, err := DisableUserByID(c, id.Hex())
	if err != nil || user.ID != id.Hex() {
		t.Fatalf("got %+v, %v", user, err)
	}

	disabled := db.Coll(model.DisabledUserCollection).LastCall().Args[0].(*model.DisabledUser)
	if disabled.UserID != id || disabled.Username != "john" {
		t.Errorf("disabled user = %+v", disabled)
	}
	checkCall(t, db, model.UserCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	// Errors: the user is kept if it can not be disabled.
	db = mock.NewDatabase("app")
	c, _ = plugintest.NewContext(db, "GET", "/", nil)
	if _, err := DisableUserByID(c, id.Hex()); err == nil {
		t.Error("an unknown user must be reported")
	}

	db = mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
	db.Coll(model.DisabledUserCollection).Err = errBoom
	c, _ = plugintest.NewContext(db, "GET", "/", nil)
	if _, err := DisableUserByID(c, id.Hex()); err == nil {
		t.Error("a failure to save the disabled user must be reported")
	}
	if got := calls(db, model.UserCollection); len(got) != 1 {
		t.Errorf("calls = %v: the user must not be deleted", got)
	}
}

func TestEnableUserByID(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.DisabledUserCollection).One = mock.Result{Value: bson.M{"_id": id, "userId": userID, "username": "john", "privateKey": "key"}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	user, err := EnableUserByID(c, id.Hex())
	if err != nil || user.ID != userID.Hex() || user.Username != "john" {
		t.Fatalf("got %+v, %v", user, err)
	}

	restored := db.Coll(model.UserCollection).LastCall().Args[0].(*model.User)
	if restored.ID != userID || restored.PrivateKey != "key" {
		t.Errorf("restored user = %+v", restored)
	}
	checkCall(t, db, model.DisabledUserCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	// Errors.
	c, _ = plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
	if _, err := EnableUserByID(c, id.Hex()); err == nil {
		t.Error("an unknown disabled user must be reported")
	}

	db.Coll(model.UserCollection).Err = errBoom
	c, _ = plugintest.NewContext(db, "GET", "/", nil)
	if _, err := EnableUserByID(c, id.Hex()); err == nil {
		t.Error("a failure to restore the user must be reported")
	}
}

func TestUpdateEncryptedPasswordByUserID(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := UpdateEncryptedPasswordByUserID(c, id.Hex(), "$argon2id$hash"); err != nil {
		t.Fatal(err)
	}
	call := db.Coll(model.UserCollection).LastCall()
	if call.Method != "UpdateByID" {
		t.Errorf("call = %+v", call)
	}

	if err := UpdateEncryptedPasswordByUserID(c, "nope", "$argon2id$hash"); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}
