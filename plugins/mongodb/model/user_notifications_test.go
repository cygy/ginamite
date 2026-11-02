package model

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestUser_NotificationUpdates(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()
	user := NewUser()
	_ = user.SetID(hex)

	tests := []struct {
		name string
		call func(db *mock.Database) error
		want bson.M
	}{
		{"UpdateNotifications", func(db *mock.Database) error { return user.UpdateNotifications(ctx, []string{"a"}, db) }, bson.M{"$set": bson.M{"notifications": []string{"a"}}}},
		{"UnsubscribeNotification", func(db *mock.Database) error { return user.UnsubscribeNotification(ctx, "a", db) }, bson.M{"$pull": bson.M{"notifications": "a"}}},
		{
			"UpdateCountOfNotifications", func(db *mock.Database) error { return user.UpdateCountOfNotifications(ctx, 10, 4, db) },
			bson.M{"$set": bson.M{"notificationsStats.total": int64(10), "notificationsStats.unread": int64(4)}},
		},
		{
			"DecrementCountOfUnreadNotifications", func(db *mock.Database) error { return user.DecrementCountOfUnreadNotifications(ctx, db) },
			bson.M{"$inc": bson.M{"notificationsStats.unread": -1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			if err := tt.call(db); err != nil {
				t.Fatal(err)
			}
			checkUpdate(t, db, UserCollection, id, tt.want)
		})
	}
}

func TestUser_GenerateUnsubscribeKey(t *testing.T) {
	user := &User{ID: mustObjectID(), PrivateKey: "secret"}

	key := user.GenerateUnsubscribeKey("news")
	if want := notifications.GenerateUnsubscribeKey(user.ID.Hex(), "secret", "news"); key != want {
		t.Errorf("key = %q, want %q", key, want)
	}
	if other := (&User{ID: user.ID, PrivateKey: "other"}).GenerateUnsubscribeKey("news"); other == key {
		t.Error("the key must depend on the private key")
	}
}

func TestUserNotificationStats_Decoding(t *testing.T) {
	// UpdateCountOfNotifications stores integers: they must be decoded.
	db := mock.NewDatabase("app")
	db.Coll(UserCollection).One = mock.Result{Value: bson.M{"_id": mustObjectID(), "notificationsStats": bson.M{"total": int64(10), "unread": int32(4)}}}

	user := NewUser()
	if err := user.GetByID(context.Background(), mustObjectID().Hex(), db); err != nil || user.NotificationsStats.Total != 10 || user.NotificationsStats.Unread != 4 {
		t.Errorf("got %+v, %v", user.NotificationsStats, err)
	}
}

func TestGetUserNotificationTargetsByType(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	newDB := func() *mock.Database {
		db := mock.NewDatabase("app")
		db.Coll(AuthTokenCollection).Docs = []any{
			bson.M{"type": "ios", "token": "token-1"},
			bson.M{"type": "android", "token": "token-2"},
		}
		db.Coll(UserCollection).One = mock.Result{Value: bson.M{"_id": id, "privateInfos": bson.M{"email": "j@example.com", "isEmailValid": true}}}

		return db
	}

	t.Run("devices and email address", func(t *testing.T) {
		db := newDB()
		targets, err := GetUserNotificationTargetsByType(ctx, hex, "news", db)
		if err != nil {
			t.Fatal(err)
		}

		want := []UserNotificationTarget{{"ios", "token-1"}, {"android", "token-2"}, {notifications.TargetEmail, "j@example.com"}}
		if !equal(targets, want) {
			t.Errorf("targets = %+v, want %+v", targets, want)
		}

		pipeline := db.Coll(AuthTokenCollection).Calls()[0].Args[0].([]bson.M)
		if len(pipeline) != 3 || pipeline[0]["$match"] == nil {
			t.Errorf("pipeline = %v", pipeline)
		}
	})

	t.Run("email address not valid", func(t *testing.T) {
		db := newDB()
		db.Coll(UserCollection).One = mock.Result{Value: bson.M{"_id": id, "privateInfos": bson.M{"email": "j@example.com"}}}

		targets, err := GetUserNotificationTargetsByType(ctx, hex, "news", db)
		if err != nil || len(targets) != 2 {
			t.Errorf("got %+v, %v", targets, err)
		}
	})

	t.Run("user does not want the notification by email", func(t *testing.T) {
		db := newDB()
		db.Coll(UserCollection).One = nil

		targets, err := GetUserNotificationTargetsByType(ctx, hex, "news", db)
		if err != nil || len(targets) != 2 {
			t.Errorf("got %+v, %v", targets, err)
		}
	})

	t.Run("errors", func(t *testing.T) {
		if _, err := GetUserNotificationTargetsByType(ctx, "nope", "news", newDB()); err == nil {
			t.Error("an invalid ID must be rejected")
		}

		db := newDB()
		db.Coll(AuthTokenCollection).Err = errBoom
		if _, err := GetUserNotificationTargetsByType(ctx, hex, "news", db); !errors.Is(err, errBoom) {
			t.Errorf("aggregate err = %v", err)
		}

		db = newDB()
		db.Coll(AuthTokenCollection).Docs = []any{make(chan int)}
		if _, err := GetUserNotificationTargetsByType(ctx, hex, "news", db); err == nil {
			t.Error("a decoding failure must be reported")
		}

		db = newDB()
		db.Coll(UserCollection).One = mock.Result{Error: errBoom}
		if _, err := GetUserNotificationTargetsByType(ctx, hex, "news", db); !errors.Is(err, errBoom) {
			t.Errorf("user query err = %v", err)
		}

		db = newDB()
		db.Coll(AuthTokenCollection).CloseErr = errBoom
		if _, err := GetUserNotificationTargetsByType(ctx, hex, "news", db); err != nil {
			t.Errorf("a failure to close the cursor is only logged, got %v", err)
		}
	})
}
