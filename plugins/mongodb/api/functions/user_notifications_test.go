package functions

import (
	"testing"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

// newNotifications returns the storage of the notifications working on the database, with "en" as default locale.
func newNotifications(db *mock.Database) *Notifications {
	return NewNotifications(func() store.Database { return db }, func() string { return "en" })
}

// notificationDocument returns a notification of the user.
func notificationDocument(id, userID bson.ObjectID, read bool) bson.M {
	return bson.M{
		"_id": id, "userId": userID, "type": "news", "read": read,
		"locales": bson.M{"en": bson.M{"title": "Title", "text": "Text", "preview": "Preview"}},
	}
}

func TestNotifications_NotificationsReceivedByUser(t *testing.T) {
	userID := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.NotificationCollection).Docs = []any{model.Notification{Type: "a", Locales: map[string]model.NotificationLocale{"en": {Title: "Title"}}}}
	s := newNotifications(db)

	got, err := s.NotificationsReceivedByUser(ctx, userID.Hex(), "en")
	notifications, ok := got.([]model.OwnSummarizedNotification)
	if err != nil || !ok || len(notifications) != 1 || notifications[0].Title != "Title" || notifications[0].Type != "a" {
		t.Errorf("got %+v, %v", got, err)
	}

	if _, err := s.NotificationsReceivedByUser(ctx, "nope", "en"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	db.Coll(model.NotificationCollection).Err = errBoom
	if _, err := s.NotificationsReceivedByUser(ctx, userID.Hex(), "en"); err == nil {
		t.Error("a failing query must be reported")
	}
}

func TestNotifications_NotificationReceivedByUser(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()

	db := mock.NewDatabase("app")
	db.Coll(model.NotificationCollection).One = mock.Result{Value: notificationDocument(id, userID, false)}
	s := newNotifications(db)

	got, err := s.NotificationReceivedByUser(ctx, userID.Hex(), id.Hex(), "en")
	if notification, ok := got.(model.OwnNotification); err != nil || !ok || notification.Title != "Title" || notification.Text != "Text" {
		t.Errorf("got %+v, %v", got, err)
	}

	// The notification of another user is not found.
	if _, err := s.NotificationReceivedByUser(ctx, bson.NewObjectID().Hex(), id.Hex(), "en"); !commonErrors.IsNotFound(err) {
		t.Errorf("err = %v", err)
	}

	for _, ids := range [][2]string{{"nope", id.Hex()}, {userID.Hex(), "nope"}} {
		if _, err := s.NotificationReceivedByUser(ctx, ids[0], ids[1], "en"); err == nil {
			t.Errorf("the IDs %v must be rejected", ids)
		}
	}

	if _, err := newNotifications(mock.NewDatabase("app")).NotificationReceivedByUser(ctx, userID.Hex(), id.Hex(), "en"); err == nil {
		t.Error("an unknown notification must be reported")
	}
}

func TestNotifications_UpdateUserNotificationRead(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()

	newStorage := func(read bool) (*mock.Database, *Notifications) {
		db := mock.NewDatabase("app")
		db.Coll(model.NotificationCollection).One = mock.Result{Value: notificationDocument(id, userID, read)}

		return db, newNotifications(db)
	}

	t.Run("unread notification", func(t *testing.T) {
		db, s := newStorage(false)
		if err := s.UpdateUserNotificationRead(ctx, userID.Hex(), id.Hex()); err != nil {
			t.Fatal(err)
		}

		checkUpdate(t, db, model.NotificationCollection, id, bson.M{"$set": bson.M{"read": true}})
		checkUpdate(t, db, model.UserCollection, userID, bson.M{"$inc": bson.M{"notificationsStats.unread": -1}})
	})

	t.Run("already read", func(t *testing.T) {
		db, s := newStorage(true)
		if err := s.UpdateUserNotificationRead(ctx, userID.Hex(), id.Hex()); err != nil {
			t.Fatal(err)
		}
		if got := calls(db, model.UserCollection); len(got) != 0 {
			t.Errorf("calls = %v", got)
		}
	})

	t.Run("errors", func(t *testing.T) {
		_, s := newStorage(false)
		if err := s.UpdateUserNotificationRead(ctx, "nope", id.Hex()); err == nil {
			t.Error("an invalid user ID must be rejected")
		}

		db, s := newStorage(false)
		db.Coll(model.NotificationCollection).Errs = map[string]error{"UpdateByID": errBoom}
		if err := s.UpdateUserNotificationRead(ctx, userID.Hex(), id.Hex()); err == nil {
			t.Error("a failure to update the notification must be reported")
		}

		// The count is only logged.
		db, s = newStorage(false)
		db.Coll(model.UserCollection).Err = errBoom
		if err := s.UpdateUserNotificationRead(ctx, userID.Hex(), id.Hex()); err != nil {
			t.Errorf("err = %v", err)
		}
	})
}

func TestNotifications_UpdateUserNotificationNotified(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()

	doc := notificationDocument(id, userID, false)
	db := mock.NewDatabase("app")
	db.Coll(model.NotificationCollection).One = mock.Result{Value: doc}

	if err := newNotifications(db).UpdateUserNotificationNotified(ctx, userID.Hex(), id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.NotificationCollection, id, bson.M{"$set": bson.M{"notified": true}})

	// Already notified: nothing to update.
	doc["notified"] = true
	db = mock.NewDatabase("app")
	db.Coll(model.NotificationCollection).One = mock.Result{Value: doc}
	s := newNotifications(db)
	if err := s.UpdateUserNotificationNotified(ctx, userID.Hex(), id.Hex()); err != nil {
		t.Fatal(err)
	}
	if got := calls(db, model.NotificationCollection); len(got) != 1 {
		t.Errorf("calls = %v, want only the lookup", got)
	}

	if err := s.UpdateUserNotificationNotified(ctx, "nope", id.Hex()); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestNotifications_DeleteUserNotification(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()

	newStorage := func() (*mock.Database, *Notifications) {
		db := mock.NewDatabase("app")
		db.Coll(model.NotificationCollection).One = mock.Result{Value: notificationDocument(id, userID, false)}
		db.Coll(model.NotificationCollection).Count = 4

		return db, newNotifications(db)
	}

	db, s := newStorage()
	if err := s.DeleteUserNotification(ctx, userID.Hex(), id.Hex()); err != nil {
		t.Fatal(err)
	}

	// The counts of the user are refreshed.
	update := db.Coll(model.UserCollection).LastCall()
	if set := update.Args[1].(bson.M)["$set"].(bson.M); update.Args[0] != userID || set["notificationsStats.total"] != int64(4) {
		t.Errorf("update = %+v", update)
	}

	if err := s.DeleteUserNotification(ctx, "nope", id.Hex()); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	for _, method := range []string{"DeleteOne", "CountDocuments"} {
		db, s = newStorage()
		db.Coll(model.NotificationCollection).Errs = map[string]error{method: errBoom}
		if err := s.DeleteUserNotification(ctx, userID.Hex(), id.Hex()); err == nil {
			t.Errorf("%s failing must be reported", method)
		}
	}

	// The count of the user is only logged.
	db, s = newStorage()
	db.Coll(model.UserCollection).Err = errBoom
	if err := s.DeleteUserNotification(ctx, userID.Hex(), id.Hex()); err != nil {
		t.Errorf("err = %v", err)
	}
}

func TestNotifications_Subscription(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "notifications": []string{"news", "offers"}}}
	s := newNotifications(db)

	if err := s.SubscribeNotificationsToUser(ctx, id.Hex(), []string{"news"}); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$set": bson.M{"notifications": []string{"news"}}})

	if err := s.SubscribeNotificationsToUser(ctx, "nope", nil); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	if got, err := s.NotificationsSubscribedByUser(ctx, id.Hex()); err != nil || len(got) != 2 || got[0] != "news" {
		t.Errorf("subscribed = %v, %v", got, err)
	}

	t.Run("unknown user", func(t *testing.T) {
		got, err := newNotifications(mock.NewDatabase("app")).NotificationsSubscribedByUser(ctx, id.Hex())
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("subscribed = %v, %v, want an empty list", got, err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).Err = errBoom

		if _, err := newNotifications(db).NotificationsSubscribedByUser(ctx, id.Hex()); err == nil {
			t.Error("a failing query must be reported")
		}
	})
}

func TestNotifications_UnsubscribeNotificationFromUser(t *testing.T) {
	id := bson.NewObjectID()
	key := notifications.GenerateUnsubscribeKey(id.Hex(), "private", "news")

	newStorage := func() (*mock.Database, *Notifications) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "privateKey": "private"}}

		return db, newNotifications(db)
	}

	db, s := newStorage()
	if err := s.UnsubscribeNotificationFromUser(ctx, id.Hex(), "news", key); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$pull": bson.M{"notifications": "news"}})

	// A key for another notification, another user or another private key is rejected.
	for _, wrong := range []string{"", "wrong", notifications.GenerateUnsubscribeKey(id.Hex(), "private", "offers"), notifications.GenerateUnsubscribeKey(id.Hex(), "other", "news")} {
		db, s = newStorage()
		if err := s.UnsubscribeNotificationFromUser(ctx, id.Hex(), "news", wrong); err == nil {
			t.Errorf("the key %q must be rejected", wrong)
		}
		if got := calls(db, model.UserCollection); len(got) != 1 {
			t.Errorf("calls = %v: nothing must be updated", got)
		}
	}

	if err := s.UnsubscribeNotificationFromUser(ctx, "nope", "news", key); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	if err := newNotifications(mock.NewDatabase("app")).UnsubscribeNotificationFromUser(ctx, id.Hex(), "news", key); err == nil {
		t.Error("an unknown user must be reported")
	}

	db, s = newStorage()
	db.Coll(model.UserCollection).Errs = map[string]error{"UpdateByID": errBoom}
	if err := s.UnsubscribeNotificationFromUser(ctx, id.Hex(), "news", key); err == nil {
		t.Error("a failure to update the user must be reported")
	}
}
