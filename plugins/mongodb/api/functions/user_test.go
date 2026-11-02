package functions

import (
	"context"
	"errors"
	"testing"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
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
			c, a := context.Background(), accountsOn(db)

			user, err := a.GetUserDetailsByIdentifier(c, tt.identifier)
			if err != nil || user.ID != id.Hex() || user.Username != "john" || user.EmailAddress != "j@example.com" {
				t.Errorf("got %+v, %v", user, err)
			}
			checkCall(t, db, model.UserCollection, "FindOne", tt.filter)
		})
	}

	c, a := context.Background(), accountsOn(mock.NewDatabase("app"))
	if _, err := a.GetUserDetailsByIdentifier(c, "nobody"); err == nil {
		t.Error("an unknown user must be reported")
	}
}

func TestGetUserDetailsByID(t *testing.T) {
	id := bson.NewObjectID()

	t.Run("from the database", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
		c, a := context.Background(), accountsOn(db)

		user, err := a.GetUserDetailsByID(c, id.Hex())
		if err != nil || user.Username != "john" {
			t.Errorf("got %+v, %v", user, err)
		}
	})

	t.Run("errors", func(t *testing.T) {
		c, a := context.Background(), accountsOn(mock.NewDatabase("app"))
		if _, err := a.GetUserDetailsByID(c, "nope"); err == nil {
			t.Error("an invalid ID must be rejected")
		}
		if _, err := a.GetUserDetailsByID(c, id.Hex()); err == nil {
			t.Error("an unknown user must be reported")
		}
	})
}

func TestGetUserDetailsBySocialNetworkID(t *testing.T) {
	id := bson.NewObjectID()

	tests := []struct {
		name   string
		call   func(a *Accounts, c context.Context) (string, error)
		filter bson.M
	}{
		{"facebook", func(a *Accounts, c context.Context) (string, error) {
			user, err := a.GetUserDetailsByFacebookUserID(c, "10157123")
			return user.Username, err
		}, bson.M{"identificationMethods.facebook.userId": "10157123"}},
		{"google", func(a *Accounts, c context.Context) (string, error) {
			user, err := a.GetUserDetailsByGoogleUserID(c, "11800")
			return user.Username, err
		}, bson.M{"identificationMethods.google.userId": "11800"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
			c, a := context.Background(), accountsOn(db)

			if username, err := tt.call(a, c); err != nil || username != "john" {
				t.Errorf("got %q, %v", username, err)
			}
			checkCall(t, db, model.UserCollection, "FindOne", tt.filter)

			db.Coll(model.UserCollection).Err = errBoom
			if _, err := tt.call(a, c); !errors.Is(err, errBoom) {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestGetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
	c, a := context.Background(), accountsOn(db)

	user, password, err := a.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(c, "john")
	if err != nil || user.ID != id.Hex() || password != "hash" {
		t.Errorf("got %+v, %q, %v", user, password, err)
	}

	a = accountsOn(mock.NewDatabase("app"))
	if _, _, err := a.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(c, "nobody"); !commonErrors.IsNotFound(err) {
		t.Errorf("err = %v", err)
	}
}

func TestDeleteUserByID(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, a := context.Background(), accountsOn(db)

	if err := a.DeleteUserByID(c, id.Hex()); err != nil {
		t.Fatal(err)
	}

	// The deletion is recorded before the user is deleted.
	deleted := db.Coll(model.DeletedUserCollection).LastCall().Args[0].(*model.DeletedUser)
	if deleted.UserID != id {
		t.Errorf("deleted user = %+v", deleted)
	}
	checkCall(t, db, model.UserCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	if err := a.DeleteUserByID(c, "nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}

	db = mock.NewDatabase("app")
	db.Coll(model.DeletedUserCollection).Err = errBoom
	a = accountsOn(db)
	if err := a.DeleteUserByID(c, id.Hex()); err == nil {
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
	c, a := context.Background(), accountsOn(db)

	user, err := a.DisableUserByID(c, id.Hex())
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
	a = accountsOn(db)
	if _, err := a.DisableUserByID(c, id.Hex()); err == nil {
		t.Error("an unknown user must be reported")
	}

	db = mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: userDocument(id)}
	db.Coll(model.DisabledUserCollection).Err = errBoom
	a = accountsOn(db)
	if _, err := a.DisableUserByID(c, id.Hex()); err == nil {
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
	c, a := context.Background(), accountsOn(db)

	user, err := a.EnableUserByID(c, id.Hex())
	if err != nil || user.ID != userID.Hex() || user.Username != "john" {
		t.Fatalf("got %+v, %v", user, err)
	}

	restored := db.Coll(model.UserCollection).LastCall().Args[0].(*model.User)
	if restored.ID != userID || restored.PrivateKey != "key" {
		t.Errorf("restored user = %+v", restored)
	}
	checkCall(t, db, model.DisabledUserCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	// Errors.
	a = accountsOn(mock.NewDatabase("app"))
	if _, err := a.EnableUserByID(c, id.Hex()); err == nil {
		t.Error("an unknown disabled user must be reported")
	}

	db.Coll(model.UserCollection).Err = errBoom
	a = accountsOn(db)
	if _, err := a.EnableUserByID(c, id.Hex()); err == nil {
		t.Error("a failure to restore the user must be reported")
	}
}

func TestUpdateEncryptedPasswordByUserID(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, a := context.Background(), accountsOn(db)

	if err := a.UpdateEncryptedPasswordByUserID(c, id.Hex(), "$argon2id$hash"); err != nil {
		t.Fatal(err)
	}
	call := db.Coll(model.UserCollection).LastCall()
	if call.Method != "UpdateByID" {
		t.Errorf("call = %+v", call)
	}

	if err := a.UpdateEncryptedPasswordByUserID(c, "nope", "$argon2id$hash"); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}
