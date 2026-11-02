package functions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestSanitizeAccounts(t *testing.T) {
	fns, db := withMainSession(t)

	sanitize := fns.SanitizeAccounts(func(got store.Database) uint {
		if got != store.Database(db) {
			t.Errorf("database = %v", got)
		}

		return 3
	})
	if got := sanitize(); got != 3 {
		t.Errorf("sanitize() = %d", got)
	}

	if got := fns.SanitizeAccounts(nil)(); got != 0 {
		t.Errorf("sanitize() = %d without function", got)
	}
}

// userTokens gives the user two auth tokens.
func userTokens(db *mock.Database, expires ...time.Time) []bson.ObjectID {
	ids := make([]bson.ObjectID, len(expires))
	docs := make([]any, len(expires))
	for i, expiresAt := range expires {
		ids[i] = bson.NewObjectID()
		docs[i] = model.AuthToken{ID: ids[i], ExpiresAt: expiresAt}
	}
	db.Coll(model.AuthTokenCollection).Docs = docs

	return ids
}

func TestDeleteUserByID(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID()
	expires := time.Now().Add(time.Hour).Truncate(time.Millisecond)

	fns, db := withMainSession(t)
	ids := userTokens(db, expires, expires.Add(time.Hour))

	var updated string
	revoked, err := fns.DeleteUserByID(func(_ context.Context, id string, got store.Database) error {
		updated = id
		if got != store.Database(db) {
			t.Errorf("database = %v", got)
		}

		return nil
	})(ctx, userID.Hex())
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if updated != userID.Hex() {
		t.Errorf("the documents referencing the user were updated for %q", updated)
	}
	if len(revoked) != 2 || !revoked[ids[0].Hex()].Equal(expires) || !revoked[ids[1].Hex()].Equal(expires.Add(time.Hour)) {
		t.Errorf("revoked = %v", revoked)
	}

	// The record of the deletion, the user and the auth tokens are deleted.
	for collection, method := range map[string]string{
		model.DeletedUserCollection: "DeleteMany", model.UserCollection: "DeleteOne", model.AuthTokenCollection: "DeleteMany",
	} {
		if call := db.Coll(collection).LastCall(); call.Method != method {
			t.Errorf("%s: last call = %+v, want %s", collection, call, method)
		}
	}

	// A user already deleted is not a failure: the function can be called again.
	fns, db = withMainSession(t)
	db.Coll(model.UserCollection).Errs = map[string]error{"DeleteOne": mongo.ErrNoDocuments}
	userTokens(db, expires)

	if revoked, err := fns.DeleteUserByID(nil)(ctx, userID.Hex()); err != nil || len(revoked) != 1 {
		t.Errorf("revoked = %v, error = %v", revoked, err)
	}
}

func TestDeleteUserByID_Failures(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID().Hex()
	expires := time.Now().Add(time.Hour)

	tests := []struct {
		name    string
		prepare func(db *mock.Database)
		stopped string // The collection which must not be touched after the failure.
	}{
		{"deletion record", func(db *mock.Database) { db.Coll(model.DeletedUserCollection).Err = errBoom }, model.UserCollection},
		{"user", func(db *mock.Database) { db.Coll(model.UserCollection).Err = errBoom }, model.AuthTokenCollection},
		{"read of the auth tokens", func(db *mock.Database) { db.Coll(model.AuthTokenCollection).Errs = map[string]error{"Find": errBoom} }, ""},
		{"deletion of the auth tokens", func(db *mock.Database) {
			db.Coll(model.AuthTokenCollection).Errs = map[string]error{"DeleteMany": errBoom}
		}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fns, db := withMainSession(t)
			userTokens(db, expires)
			tt.prepare(db)

			revoked, err := fns.DeleteUserByID(nil)(ctx, userID)

			if !errors.Is(err, errBoom) || revoked != nil {
				t.Errorf("revoked = %v, error = %v, want the error of the database", revoked, err)
			}
			if tt.stopped != "" && len(db.Coll(tt.stopped).Calls()) != 0 {
				t.Errorf("%s was touched after the failure: %v", tt.stopped, methods(db, tt.stopped))
			}
		})
	}

	t.Run("tokens not read are not deleted", func(t *testing.T) {
		// Their revocation could not be told to the other servers: the deletion is tried again.
		fns, db := withMainSession(t)
		db.Coll(model.AuthTokenCollection).Errs = map[string]error{"Find": errBoom}

		if _, err := fns.DeleteUserByID(nil)(ctx, userID); err == nil {
			t.Fatal("error = nil")
		}
		for _, method := range methods(db, model.AuthTokenCollection) {
			if method == "DeleteMany" {
				t.Error("the auth tokens must not be deleted")
			}
		}
	})
}

func TestDisableUserByID(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID()
	expires := time.Now().Add(time.Hour).Truncate(time.Millisecond)

	fns, db := withMainSession(t)
	ids := userTokens(db, expires)

	var updated string
	revoked, err := fns.DisableUserByID(func(_ context.Context, id string, _ store.Database) error { updated = id; return nil })(ctx, userID.Hex())

	if err != nil || updated != userID.Hex() || len(revoked) != 1 || !revoked[ids[0].Hex()].Equal(expires) {
		t.Errorf("updated = %q, revoked = %v, error = %v", updated, revoked, err)
	}
	if call := db.Coll(model.UserCollection).LastCall(); call.Method != "DeleteOne" {
		t.Errorf("user call = %+v", call)
	}
	if got := methods(db, model.DeletedUserCollection); len(got) != 0 {
		t.Errorf("the deletion must not be recorded for a disabled user: %v", got)
	}

	fns, db = withMainSession(t)
	db.Coll(model.UserCollection).Err = errBoom
	if revoked, err := fns.DisableUserByID(nil)(ctx, userID.Hex()); !errors.Is(err, errBoom) || len(revoked) != 0 {
		t.Errorf("revoked = %v, error = %v", revoked, err)
	}

	// A user already deleted is not a failure.
	fns, db = withMainSession(t)
	db.Coll(model.UserCollection).Errs = map[string]error{"DeleteOne": mongo.ErrNoDocuments}
	if _, err := fns.DisableUserByID(nil)(ctx, userID.Hex()); err != nil {
		t.Errorf("error = %v", err)
	}
}

func TestEnableUserByID(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID()

	fns, db := withMainSession(t)
	var updated string
	if err := fns.EnableUserByID(func(_ context.Context, id string, _ store.Database) error { updated = id; return nil })(ctx, userID.Hex()); err != nil {
		t.Errorf("error = %v", err)
	}

	if updated != userID.Hex() {
		t.Errorf("updated = %q", updated)
	}
	if call := db.Coll(model.DisabledUserCollection).LastCall(); call.Method != "DeleteMany" || call.Args[0].(bson.M)["userId"] != userID {
		t.Errorf("disabled user call = %+v", call)
	}

	fns, db = withMainSession(t)
	db.Coll(model.DisabledUserCollection).Err = errBoom
	if err := fns.EnableUserByID(nil)(ctx, userID.Hex()); !errors.Is(err, errBoom) {
		t.Errorf("error = %v", err)
	}
}

func TestUpdateUserSocialNetworksByID(t *testing.T) {
	fns, _ := withMainSession(t)

	var updated string
	if err := fns.UpdateUserSocialNetworksByID(func(_ context.Context, id string, _ store.Database) error { updated = id; return nil })(context.Background(), "user-1"); err != nil {
		t.Errorf("error = %v", err)
	}
	if updated != "user-1" {
		t.Errorf("updated = %q", updated)
	}

	// Nothing to do without function.
	if err := fns.UpdateUserSocialNetworksByID(nil)(context.Background(), "user-1"); err != nil {
		t.Errorf("error = %v", err)
	}
}

func TestGetNotificationTargetsByUserAndType(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID()

	fns, db := withMainSession(t)
	db.Coll(model.AuthTokenCollection).Docs = []any{bson.M{"type": "ios", "token": "token-1"}}
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": userID, "privateInfos": bson.M{"email": "j@example.com", "isEmailValid": true}}}

	targets, err := fns.GetNotificationTargetsByUserAndType(ctx, userID.Hex(), "news")
	want := []notifications.Target{{Target: "ios", Value: "token-1"}, {Target: notifications.TargetEmail, Value: "j@example.com"}}
	if err != nil || len(targets) != 2 || targets[0] != want[0] || targets[1] != want[1] {
		t.Errorf("got %+v, %v", targets, err)
	}

	if targets, err := fns.GetNotificationTargetsByUserAndType(ctx, "nope", "news"); err == nil || targets == nil || len(targets) != 0 {
		t.Errorf("an invalid user ID must fail with an empty list: %+v, %v", targets, err)
	}
}

func TestIsEmailAddressValid(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		count      int64
		want       bool
	}{
		{"enabled user", model.UserCollection, 1, true},
		{"disabled user", model.DisabledUserCollection, 1, true},
		{"nobody", model.UserCollection, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fns, db := withMainSession(t)
			db.Coll(tt.collection).Count = tt.count

			if got, err := fns.IsEmailAddressValid(ctx, "j@example.com"); err != nil || got != tt.want {
				t.Errorf("got %v, %v", got, err)
			}
		})
	}

	for _, collection := range []string{model.UserCollection, model.DisabledUserCollection} {
		fns, db := withMainSession(t)
		db.Coll(collection).Err = errBoom

		if got, err := fns.IsEmailAddressValid(ctx, "j@example.com"); err == nil || got {
			t.Errorf("%s failing: got %v, %v", collection, got, err)
		}
	}
}

func TestGetAccounts(t *testing.T) {
	ctx := context.Background()
	first, second := bson.NewObjectID(), bson.NewObjectID()

	fns, db := withMainSession(t)
	db.Coll(model.UserCollection).Docs = []any{model.User{ID: first}, model.User{ID: second}}

	// The inactive accounts come last: the following checks are about the last query.
	for _, tt := range []struct {
		name string
		get  func() []string
	}{
		{"never used", func() []string { return fns.GetNeverUsedAccounts(ctx, 30) }},
		{"inactive", func() []string { return fns.GetInactiveAccounts(ctx, 6) }},
	} {
		if ids := tt.get(); len(ids) != 2 || ids[0] != first.Hex() || ids[1] != second.Hex() {
			t.Errorf("%s: ids = %v", tt.name, ids)
		}
	}

	// The intervals are in days and in months of 30 days.
	filter := db.Coll(model.UserCollection).LastCall().Args[0].(bson.D)
	limit := filter[0].Value.(bson.M)["$lt"].(time.Time)
	if got := time.Since(limit); got < 179*24*time.Hour || got > 181*24*time.Hour {
		t.Errorf("the inactive limit is %v ago, want 180 days", got)
	}

	// A failing query is logged and gives no account.
	db.Coll(model.UserCollection).Err = errBoom
	if ids := fns.GetNeverUsedAccounts(ctx, 30); len(ids) != 0 {
		t.Errorf("ids = %v", ids)
	}
	if ids := fns.GetInactiveAccounts(ctx, 6); len(ids) != 0 {
		t.Errorf("ids = %v", ids)
	}
}

func TestIsEmptyUser(t *testing.T) {
	fns, db := withMainSession(t)

	isEmpty := fns.IsEmptyUser(func(_ context.Context, id string, got store.Database) bool {
		return id == "empty" && got == store.Database(db)
	})
	if !isEmpty(context.Background(), "empty") || isEmpty(context.Background(), "used") {
		t.Error("fns.IsEmptyUser() does not follow the function")
	}

	if !fns.IsEmptyUser(nil)(context.Background(), "anyone") {
		t.Error("every user is empty without function")
	}
}

func TestUsers_HookErrors(t *testing.T) {
	fns, _ := withMainSession(t)
	ctx := context.Background()
	userID := bson.NewObjectID().Hex()
	failing := func(context.Context, string, store.Database) error { return errBoom }

	if revoked, err := fns.DeleteUserByID(failing)(ctx, userID); !errors.Is(err, errBoom) || revoked != nil {
		t.Errorf("DeleteUserByID() = %v, %v", revoked, err)
	}
	if revoked, err := fns.DisableUserByID(failing)(ctx, userID); !errors.Is(err, errBoom) || revoked != nil {
		t.Errorf("DisableUserByID() = %v, %v", revoked, err)
	}
	if err := fns.EnableUserByID(failing)(ctx, userID); !errors.Is(err, errBoom) {
		t.Errorf("EnableUserByID() = %v", err)
	}
	if err := fns.UpdateUserSocialNetworksByID(failing)(ctx, userID); !errors.Is(err, errBoom) {
		t.Errorf("UpdateUserSocialNetworksByID() = %v", err)
	}
}
