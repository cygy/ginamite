package model

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store/mock"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestUser_Document(t *testing.T) {
	testDocument(t, UserCollection, func() document.Document { return NewUser() })
}

func TestNewUser(t *testing.T) {
	if user := NewUser(); user.IsAdmin || user.IsAnonymous || user.IsSaved() {
		t.Errorf("user = %+v", user)
	}
}

func TestUserWithID(t *testing.T) {
	id, hex := savedID()

	user, err := UserWithID(hex)
	if err != nil || user.ID != id {
		t.Errorf("got %+v, %v", user, err)
	}
	if _, err := UserWithID("nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}

func TestUser_SetAdmin(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()
	user := NewUser()
	_ = user.SetID(hex)

	defer func(old *config.Configuration) { config.Main = old }(config.Main)

	// Without configuration, or in the test environment, the flag is saved.
	for name, conf := range map[string]*config.Configuration{"no configuration": nil, "test environment": {Environment: "test"}} {
		t.Run(name, func(t *testing.T) {
			config.Main = conf
			db := mock.NewDatabase("app")

			if err := user.SetAdmin(ctx, true, db); err != nil {
				t.Fatal(err)
			}
			checkUpdate(t, db, UserCollection, id, bson.M{"$set": bson.M{"isAdmin": true}})
		})
	}

	// Anywhere else it is only meant for the tests: nothing is saved.
	t.Run("production", func(t *testing.T) {
		config.Main = &config.Configuration{Environment: "production"}
		db := mock.NewDatabase("app")

		if err := user.SetAdmin(ctx, true, db); err != nil {
			t.Fatal(err)
		}
		if calls := db.Coll(UserCollection).Calls(); len(calls) != 0 {
			t.Errorf("calls = %+v", calls)
		}
	})
}

func TestUser_Finders(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	tests := []struct {
		name   string
		call   func(user *User, db *mock.Database) error
		filter any
	}{
		{"username", func(u *User, db *mock.Database) error { return u.GetByUsername(ctx, "john", db) }, bson.M{"username": "john"}},
		{"email address", func(u *User, db *mock.Database) error { return u.GetByEmailAddress(ctx, "j@example.com", db) }, bson.M{"privateInfos.email": "j@example.com"}},
		{"facebook ID", func(u *User, db *mock.Database) error { return u.GetByFacebookUserID(ctx, "123456", db) }, bson.M{"identificationMethods.facebook.userId": "123456"}},
		{"google ID", func(u *User, db *mock.Database) error { return u.GetByGoogleUserID(ctx, "987", db) }, bson.M{"identificationMethods.google.userId": "987"}},
		{
			"ID and notification", func(u *User, db *mock.Database) error { return u.GetByIDAndNotification(ctx, hex, "news", db) },
			bson.D{{Key: "_id", Value: id}, {Key: "notifications", Value: bson.M{"$in": []string{"news"}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll(UserCollection).One = mock.Result{Value: bson.M{"_id": id, "username": "john"}}

			user := NewUser()
			if err := tt.call(user, db); err != nil || user.Username != "john" {
				t.Fatalf("got %+v, %v", user, err)
			}
			checkCall(t, db, UserCollection, "FindOne", tt.filter)

			db.Coll(UserCollection).Err = errBoom
			if err := tt.call(NewUser(), db); !errors.Is(err, errBoom) {
				t.Errorf("err = %v", err)
			}
		})
	}

	if err := NewUser().GetByIDAndNotification(ctx, "nope", "news", mock.NewDatabase("app")); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}

func TestUser_GetByIdentifier(t *testing.T) {
	ctx := context.Background()
	id := mustObjectID()

	t.Run("username", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(UserCollection).One = mock.Result{Value: bson.M{"_id": id, "username": "john"}}

		user := NewUser()
		if err := user.GetByIdentifier(ctx, "john", db); err != nil || user.ID != id {
			t.Errorf("got %+v, %v", user, err)
		}
		if n := len(db.Coll(UserCollection).Calls()); n != 1 {
			t.Errorf("%d queries, want only the one by username", n)
		}
	})

	t.Run("email address", func(t *testing.T) {
		db := mock.NewDatabase("app")
		queries := 0
		db.Coll(UserCollection).One = funcResult(func(v any) error {
			queries++
			if queries == 1 {
				return mongo.ErrNoDocuments
			}

			return mock.Result{Value: bson.M{"_id": id}}.Decode(v)
		})

		user := NewUser()
		if err := user.GetByIdentifier(ctx, "j@example.com", db); err != nil || user.ID != id {
			t.Errorf("got %+v, %v", user, err)
		}
		if got := lastCall(db, UserCollection).Args[0]; !equal(got, bson.D{{Key: "privateInfos.email", Value: "j@example.com"}}) {
			t.Errorf("filter = %v", got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := NewUser().GetByIdentifier(ctx, "nobody", db); !commonErrors.IsNotFound(err) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("errors", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(UserCollection).Err = errBoom
		if err := NewUser().GetByIdentifier(ctx, "john", db); !errors.Is(err, errBoom) {
			t.Errorf("username query err = %v", err)
		}

		// The error of the second query is reported too.
		db = mock.NewDatabase("app")
		queries := 0
		db.Coll(UserCollection).One = funcResult(func(any) error {
			queries++
			if queries == 1 {
				return mongo.ErrNoDocuments
			}

			return errBoom
		})
		if err := NewUser().GetByIdentifier(ctx, "john", db); !errors.Is(err, errBoom) {
			t.Errorf("email query err = %v", err)
		}
	})
}

func TestUser_UpdateLastLogin(t *testing.T) {
	id, hex := savedID()
	user := NewUser()
	_ = user.SetID(hex)

	db := mock.NewDatabase("app")
	before := time.Now()
	if err := user.UpdateLastLogin(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	set := lastCall(db, UserCollection).Args[1].(bson.M)["$set"].(bson.M)["lastLogin"].(time.Time)
	if lastCall(db, UserCollection).Args[0] != id || set.Before(before) {
		t.Errorf("call = %+v", lastCall(db, UserCollection))
	}
}

func TestUser_ToUser(t *testing.T) {
	user := &User{
		ID: mustObjectID(), Username: "john", PrivateKey: "key",
		PrivateInfos: UserPrivateInfos{Email: "j@example.com", IsEmailValid: true},
		Settings:     UserSettings{Locale: "fr"},
	}

	want := authentication.User{ID: user.ID.Hex(), Username: "john", EmailAddress: "j@example.com", Locale: "fr", PrivateKey: "key", IsEmailValid: true}
	if got := user.ToUser(); got != want {
		t.Errorf("got %+v", got)
	}
}

func TestIsUserExisting(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		call   func(db *mock.Database) (bool, error)
		filter any
	}{
		{"username", func(db *mock.Database) (bool, error) { return IsUserWithUsernameExisting(ctx, "john", db) }, bson.M{"username": "john"}},
		{
			"email address", func(db *mock.Database) (bool, error) { return IsUserWithEmailAddressExisting(ctx, "j@example.com", db) },
			bson.M{"privateInfos.email": "j@example.com"},
		},
		{
			"valid email address", func(db *mock.Database) (bool, error) {
				return IsUserWithValidEmailAddressExisting(ctx, "j@example.com", db)
			},
			bson.M{"privateInfos.email": "j@example.com", "privateInfos.isEmailValid": true},
		},
		// The IDs of the social networks are not ObjectIDs.
		{
			"facebook ID", func(db *mock.Database) (bool, error) {
				return IsUserWithFacebookUserIDExisting(ctx, "10157123456789", db)
			},
			bson.M{"identificationMethods.facebook.userId": "10157123456789"},
		},
		{
			"google ID", func(db *mock.Database) (bool, error) { return IsUserWithGoogleUserIDExisting(ctx, "1180000000000", db) },
			bson.M{"identificationMethods.google.userId": "1180000000000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			if got, err := tt.call(db); err != nil || got {
				t.Errorf("without document: got %v, %v", got, err)
			}
			checkCall(t, db, UserCollection, "CountDocuments", tt.filter)

			db.Coll(UserCollection).Count = 2
			if got, err := tt.call(db); err != nil || !got {
				t.Errorf("with documents: got %v, %v", got, err)
			}

			db.Coll(UserCollection).Err = errBoom
			if _, err := tt.call(db); !errors.Is(err, errBoom) {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestDeleteUserByUserID(t *testing.T) {
	id, hex := savedID()

	db := mock.NewDatabase("app")
	if err := DeleteUserByUserID(context.Background(), hex, db); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, UserCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	if err := DeleteUserByUserID(context.Background(), "nope", db); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}

func TestGetAnonymousUser(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	db := mock.NewDatabase("app")
	db.Coll(UserCollection).One = mock.Result{Value: bson.M{"_id": id, "isAnonymous": true}}

	user, err := GetAnonymousUser(ctx, db)
	if err != nil || user.ID != id || !user.IsAnonymous {
		t.Errorf("got %+v, %v", user, err)
	}
	checkCall(t, db, UserCollection, "FindOne", bson.M{"isAnonymous": true})

	if got, err := GetAnonymousUserID(ctx, db); err != nil || got != hex {
		t.Errorf("got %q, %v", got, err)
	}

	db.Coll(UserCollection).Err = errBoom
	if _, err := GetAnonymousUser(ctx, db); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
	if _, err := GetAnonymousUserID(ctx, db); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestGetUserIDs(t *testing.T) {
	ctx := context.Background()
	first, second := mustObjectID(), mustObjectID()

	db := mock.NewDatabase("app")
	db.Coll(UserCollection).Docs = []any{User{ID: first}, User{ID: second}}

	tests := []struct {
		name string
		call func() ([]string, error)
		key  string
	}{
		{"never used", func() ([]string, error) { return GetNeverUsedUserIDs(ctx, time.Hour, db) }, "registrationInfos.date"},
		{"inactive", func() ([]string, error) { return GetInactiveUserIDs(ctx, time.Hour, db) }, "lastLogin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			ids, err := tt.call()
			if err != nil || len(ids) != 2 || ids[0] != first.Hex() || ids[1] != second.Hex() {
				t.Fatalf("got %v, %v", ids, err)
			}

			filter := lastCall(db, UserCollection).Args[0].(bson.D)
			if filter[0].Key != tt.key {
				t.Fatalf("filter = %v", filter)
			}

			// The date is one interval ago.
			limit := filter[0].Value.(bson.M)["$lt"].(time.Time)
			if limit.After(before) || limit.Before(before.Add(-time.Hour-time.Minute)) {
				t.Errorf("limit = %v, want about an hour ago", limit)
			}
		})
	}

	db.Coll(UserCollection).Err = errBoom
	if ids, err := GetInactiveUserIDs(ctx, time.Hour, db); !errors.Is(err, errBoom) || len(ids) != 0 {
		t.Errorf("got %v, %v", ids, err)
	}
}
