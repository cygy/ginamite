package functions

import (
	"context"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

func TestSanitizeAccounts(t *testing.T) {
	db := withMainSession(t)

	sanitize := SanitizeAccounts(func(got store.Database) uint {
		if got != store.Database(db) {
			t.Errorf("database = %v", got)
		}

		return 3
	})
	if got := sanitize(); got != 3 {
		t.Errorf("sanitize() = %d", got)
	}

	if got := SanitizeAccounts(nil)(); got != 0 {
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

	db := withMainSession(t)
	ids := userTokens(db, expires, expires.Add(time.Hour))

	var updated string
	revoked := DeleteUserByID(func(_ context.Context, id string, got store.Database) {
		updated = id
		if got != store.Database(db) {
			t.Errorf("database = %v", got)
		}
	})(ctx, userID.Hex())

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

	// Without function to update the documents, and with failing queries, everything else is done.
	db = withMainSession(t)
	db.Coll(model.DeletedUserCollection).Err = errBoom
	db.Coll(model.UserCollection).Err = errBoom
	db.Coll(model.AuthTokenCollection).Errs = map[string]error{"DeleteMany": errBoom}
	userTokens(db, expires)

	if revoked := DeleteUserByID(nil)(ctx, userID.Hex()); len(revoked) != 1 {
		t.Errorf("revoked = %v", revoked)
	}
	if got := methods(db, model.AuthTokenCollection); len(got) != 2 {
		t.Errorf("auth token calls = %v", got)
	}
}

func TestDeleteUserByID_TokensNotRead(t *testing.T) {
	// Failing to read the auth tokens gives no revoked token, but the tokens are still deleted.
	db := withMainSession(t)
	db.Coll(model.AuthTokenCollection).Errs = map[string]error{"Find": errBoom}

	revoked := DeleteUserByID(nil)(context.Background(), bson.NewObjectID().Hex())
	if revoked == nil || len(revoked) != 0 {
		t.Errorf("revoked = %v", revoked)
	}
	if got := methods(db, model.AuthTokenCollection); len(got) != 2 || got[1] != "DeleteMany" {
		t.Errorf("auth token calls = %v", got)
	}
}

func TestDisableUserByID(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID()
	expires := time.Now().Add(time.Hour).Truncate(time.Millisecond)

	db := withMainSession(t)
	ids := userTokens(db, expires)

	var updated string
	revoked := DisableUserByID(func(_ context.Context, id string, _ store.Database) { updated = id })(ctx, userID.Hex())

	if updated != userID.Hex() || len(revoked) != 1 || !revoked[ids[0].Hex()].Equal(expires) {
		t.Errorf("updated = %q, revoked = %v", updated, revoked)
	}
	if call := db.Coll(model.UserCollection).LastCall(); call.Method != "DeleteOne" {
		t.Errorf("user call = %+v", call)
	}
	if got := methods(db, model.DeletedUserCollection); len(got) != 0 {
		t.Errorf("the deletion must not be recorded for a disabled user: %v", got)
	}

	db = withMainSession(t)
	db.Coll(model.UserCollection).Err = errBoom
	if revoked := DisableUserByID(nil)(ctx, userID.Hex()); len(revoked) != 0 {
		t.Errorf("revoked = %v", revoked)
	}
}

func TestEnableUserByID(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID()

	db := withMainSession(t)
	var updated string
	EnableUserByID(func(_ context.Context, id string, _ store.Database) { updated = id })(ctx, userID.Hex())

	if updated != userID.Hex() {
		t.Errorf("updated = %q", updated)
	}
	if call := db.Coll(model.DisabledUserCollection).LastCall(); call.Method != "DeleteMany" || call.Args[0].(bson.M)["userId"] != userID {
		t.Errorf("disabled user call = %+v", call)
	}

	// Failures are only logged.
	db = withMainSession(t)
	db.Coll(model.DisabledUserCollection).Err = errBoom
	EnableUserByID(nil)(ctx, userID.Hex())
}

func TestUpdateUserSocialNetworksByID(t *testing.T) {
	withMainSession(t)

	var updated string
	UpdateUserSocialNetworksByID(func(_ context.Context, id string, _ store.Database) { updated = id })(context.Background(), "user-1")
	if updated != "user-1" {
		t.Errorf("updated = %q", updated)
	}

	// Nothing to do without function.
	UpdateUserSocialNetworksByID(nil)(context.Background(), "user-1")
}

func TestGetNotificationTargetsByUserAndType(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID()

	db := withMainSession(t)
	db.Coll(model.AuthTokenCollection).Docs = []any{bson.M{"type": "ios", "token": "token-1"}}
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": userID, "privateInfos": bson.M{"email": "j@example.com", "isEmailValid": true}}}

	targets, err := GetNotificationTargetsByUserAndType(ctx, userID.Hex(), "news")
	want := []notifications.Target{{Target: "ios", Value: "token-1"}, {Target: notifications.TargetEmail, Value: "j@example.com"}}
	if err != nil || len(targets) != 2 || targets[0] != want[0] || targets[1] != want[1] {
		t.Errorf("got %+v, %v", targets, err)
	}

	if targets, err := GetNotificationTargetsByUserAndType(ctx, "nope", "news"); err == nil || targets == nil || len(targets) != 0 {
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
			db := withMainSession(t)
			db.Coll(tt.collection).Count = tt.count

			if got, err := IsEmailAddressValid(ctx, "j@example.com"); err != nil || got != tt.want {
				t.Errorf("got %v, %v", got, err)
			}
		})
	}

	for _, collection := range []string{model.UserCollection, model.DisabledUserCollection} {
		db := withMainSession(t)
		db.Coll(collection).Err = errBoom

		if got, err := IsEmailAddressValid(ctx, "j@example.com"); err == nil || got {
			t.Errorf("%s failing: got %v, %v", collection, got, err)
		}
	}
}

func TestGetAccounts(t *testing.T) {
	ctx := context.Background()
	first, second := bson.NewObjectID(), bson.NewObjectID()

	db := withMainSession(t)
	db.Coll(model.UserCollection).Docs = []any{model.User{ID: first}, model.User{ID: second}}

	// The inactive accounts come last: the following checks are about the last query.
	for _, tt := range []struct {
		name string
		get  func() []string
	}{
		{"never used", func() []string { return GetNeverUsedAccounts(ctx, 30) }},
		{"inactive", func() []string { return GetInactiveAccounts(ctx, 6) }},
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
	if ids := GetNeverUsedAccounts(ctx, 30); len(ids) != 0 {
		t.Errorf("ids = %v", ids)
	}
	if ids := GetInactiveAccounts(ctx, 6); len(ids) != 0 {
		t.Errorf("ids = %v", ids)
	}
}

func TestIsEmptyUser(t *testing.T) {
	db := withMainSession(t)

	isEmpty := IsEmptyUser(func(_ context.Context, id string, got store.Database) bool {
		return id == "empty" && got == store.Database(db)
	})
	if !isEmpty(context.Background(), "empty") || isEmpty(context.Background(), "used") {
		t.Error("IsEmptyUser() does not follow the function")
	}

	if !IsEmptyUser(nil)(context.Background(), "anyone") {
		t.Error("every user is empty without function")
	}
}
