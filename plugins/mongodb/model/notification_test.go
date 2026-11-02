package model

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestNotification_Document(t *testing.T) {
	testDocument(t, NotificationCollection, func() document.Document { return NewNotification() })
}

func TestNewNotification(t *testing.T) {
	notification := NewNotification()
	if notification.Read || notification.Notified || notification.CreatedAt.IsZero() {
		t.Errorf("notification = %+v", notification)
	}
}

func TestNotification_Updates(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()
	notification := NewNotification()
	_ = notification.SetID(hex)

	db := mock.NewDatabase("app")
	if err := notification.UpdateRead(ctx, db); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, NotificationCollection, id, bson.M{"$set": bson.M{"read": true}})

	if err := notification.UpdateNotified(ctx, db); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, NotificationCollection, id, bson.M{"$set": bson.M{"notified": true}})
}

func TestGetUserNotifications(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	db := mock.NewDatabase("app")
	db.Coll(NotificationCollection).Docs = []any{Notification{Type: "a"}, Notification{Type: "b"}}

	notifications, err := GetUserNotifications(ctx, hex, db)
	if err != nil || len(notifications) != 2 || notifications[0].Type != "a" {
		t.Fatalf("got %+v, %v", notifications, err)
	}
	checkCall(t, db, NotificationCollection, "Find", bson.M{"userId": id})

	if notifications, err := GetUserNotifications(ctx, "nope", db); err == nil || notifications == nil {
		t.Errorf("an invalid ID must fail with an empty list: %v, %v", notifications, err)
	}
}

func TestGetCountOfUserNotifications(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	db := mock.NewDatabase("app")
	db.Coll(NotificationCollection).Count = 7

	if n, err := GetCountOfUserNotifications(ctx, hex, db); err != nil || n != 7 {
		t.Errorf("got %d, %v", n, err)
	}
	checkCall(t, db, NotificationCollection, "CountDocuments", bson.M{"userId": id})

	if _, err := GetCountOfUserNotifications(ctx, "nope", db); err == nil {
		t.Error("an invalid ID must be rejected")
	}

	if n, err := GetCountOfUserUnreadNotifications(ctx, hex, db); err != nil || n != 7 {
		t.Errorf("got %d, %v", n, err)
	}
	checkCall(t, db, NotificationCollection, "CountDocuments", bson.D{{Key: "userId", Value: id}, {Key: "read", Value: false}})

	if _, err := GetCountOfUserUnreadNotifications(ctx, "nope", db); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}

func TestGetOldestUserNotifications(t *testing.T) {
	ctx := context.Background()
	first, second := mustObjectID(), mustObjectID()
	_, hex := savedID()

	db := mock.NewDatabase("app")
	db.Coll(NotificationCollection).Docs = []any{Notification{ID: first}, Notification{ID: second}}

	ids, err := GetOldestUserNotifications(ctx, hex, 50, db)
	if err != nil || len(ids) != 2 || ids[0] != first || ids[1] != second {
		t.Fatalf("got %v, %v", ids, err)
	}
	if opts := mock.FindOptions(lastCall(db, NotificationCollection).Opts); *opts.Skip != 50 || opts.Limit != nil {
		t.Errorf("options = skip %v, limit %v: the offset is expected without limit", opts.Skip, opts.Limit)
	}

	if ids, err := GetOldestUserNotifications(ctx, "nope", 0, db); err == nil || ids == nil {
		t.Errorf("an invalid ID must fail with an empty list: %v, %v", ids, err)
	}

	db.Coll(NotificationCollection).Err = errBoom
	if _, err := GetOldestUserNotifications(ctx, hex, 0, db); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestDeleteUserNotifications(t *testing.T) {
	ctx := context.Background()
	first, second := mustObjectID(), mustObjectID()

	db := mock.NewDatabase("app")
	if _, err := DeleteUserNotifications(ctx, []bson.ObjectID{first, second}, db); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, NotificationCollection, "DeleteMany", bson.M{"$or": []bson.M{{"_id": first}, {"_id": second}}})

	// MongoDB rejects an empty $or: the database must not be called.
	db = mock.NewDatabase("app")
	if n, err := DeleteUserNotifications(ctx, nil, db); err != nil || n != 0 {
		t.Errorf("got %d, %v", n, err)
	}
	if calls := db.Coll(NotificationCollection).Calls(); len(calls) != 0 {
		t.Errorf("calls = %+v", calls)
	}
}
