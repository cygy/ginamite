package functions

import (
	"testing"

	apiContextLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/common/config"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// withDefaultLocale sets the default locale of the application for the duration of the test.
func withDefaultLocale(t *testing.T, locale string) {
	t.Helper()

	old := config.Main
	t.Cleanup(func() { config.Main = old })

	config.Main = config.NewConfiguration()
	config.Main.DefaultLocale = locale
}

// notificationDocument returns a notification of the user.
func notificationDocument(id, userID bson.ObjectID, read bool) bson.M {
	return bson.M{
		"_id": id, "userId": userID, "type": "news", "read": read,
		"locales": bson.M{"en": bson.M{"title": "Title", "text": "Text", "preview": "Preview"}},
	}
}

func TestNotificationsReceivedByUser(t *testing.T) {
	withDefaultLocale(t, "en")
	userID := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.NotificationCollection).Docs = []any{model.Notification{Type: "a", Locales: map[string]model.NotificationLocale{"en": {Title: "Title"}}}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)
	apiContextLocale.Set(c, "", "en")

	got, err := NotificationsReceivedByUser(c, userID.Hex())
	notifications, ok := got.([]model.OwnSummarizedNotification)
	if err != nil || !ok || len(notifications) != 1 || notifications[0].Title != "Title" || notifications[0].Type != "a" {
		t.Errorf("got %+v, %v", got, err)
	}

	if _, err := NotificationsReceivedByUser(c, "nope"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	db.Coll(model.NotificationCollection).Err = errBoom
	if _, err := NotificationsReceivedByUser(c, userID.Hex()); err == nil {
		t.Error("a failing query must be reported")
	}
}

func TestNotificationReceivedByUser(t *testing.T) {
	withDefaultLocale(t, "en")
	id, userID := bson.NewObjectID(), bson.NewObjectID()

	db := mock.NewDatabase("app")
	db.Coll(model.NotificationCollection).One = mock.Result{Value: notificationDocument(id, userID, false)}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)
	apiContextLocale.Set(c, "", "en")

	got, err := NotificationReceivedByUser(c, userID.Hex(), id.Hex())
	if notification, ok := got.(model.OwnNotification); err != nil || !ok || notification.Title != "Title" || notification.Text != "Text" {
		t.Errorf("got %+v, %v", got, err)
	}

	// The notification of another user is not found.
	if _, err := NotificationReceivedByUser(c, bson.NewObjectID().Hex(), id.Hex()); !commonErrors.IsNotFound(err) {
		t.Errorf("err = %v", err)
	}

	for _, ids := range [][2]string{{"nope", id.Hex()}, {userID.Hex(), "nope"}} {
		if _, err := NotificationReceivedByUser(c, ids[0], ids[1]); err == nil {
			t.Errorf("the IDs %v must be rejected", ids)
		}
	}

	c, _ = plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
	if _, err := NotificationReceivedByUser(c, userID.Hex(), id.Hex()); err == nil {
		t.Error("an unknown notification must be reported")
	}
}

func TestUpdateUserNotificationRead(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()

	newContext := func(read bool) (*mock.Database, *gin.Context) {
		db := mock.NewDatabase("app")
		db.Coll(model.NotificationCollection).One = mock.Result{Value: notificationDocument(id, userID, read)}
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		return db, c
	}

	t.Run("unread notification", func(t *testing.T) {
		db, c := newContext(false)
		if err := UpdateUserNotificationRead(c, userID.Hex(), id.Hex()); err != nil {
			t.Fatal(err)
		}

		checkUpdate(t, db, model.NotificationCollection, id, bson.M{"$set": bson.M{"read": true}})
		checkUpdate(t, db, model.UserCollection, userID, bson.M{"$inc": bson.M{"notificationsStats.unread": -1}})
	})

	t.Run("already read", func(t *testing.T) {
		db, c := newContext(true)
		if err := UpdateUserNotificationRead(c, userID.Hex(), id.Hex()); err != nil {
			t.Fatal(err)
		}
		if got := calls(db, model.UserCollection); len(got) != 0 {
			t.Errorf("calls = %v", got)
		}
	})

	t.Run("errors", func(t *testing.T) {
		_, c := newContext(false)
		if err := UpdateUserNotificationRead(c, "nope", id.Hex()); err == nil {
			t.Error("an invalid user ID must be rejected")
		}

		db, c := newContext(false)
		db.Coll(model.NotificationCollection).Errs = map[string]error{"UpdateByID": errBoom}
		if err := UpdateUserNotificationRead(c, userID.Hex(), id.Hex()); err == nil {
			t.Error("a failure to update the notification must be reported")
		}

		// The count is only logged.
		db, c = newContext(false)
		db.Coll(model.UserCollection).Err = errBoom
		if err := UpdateUserNotificationRead(c, userID.Hex(), id.Hex()); err != nil {
			t.Errorf("err = %v", err)
		}
	})
}

func TestUpdateUserNotificationNotified(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()

	doc := notificationDocument(id, userID, false)
	db := mock.NewDatabase("app")
	db.Coll(model.NotificationCollection).One = mock.Result{Value: doc}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := UpdateUserNotificationNotified(c, userID.Hex(), id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.NotificationCollection, id, bson.M{"$set": bson.M{"notified": true}})

	// Already notified: nothing to update.
	doc["notified"] = true
	db = mock.NewDatabase("app")
	db.Coll(model.NotificationCollection).One = mock.Result{Value: doc}
	c, _ = plugintest.NewContext(db, "GET", "/", nil)
	if err := UpdateUserNotificationNotified(c, userID.Hex(), id.Hex()); err != nil {
		t.Fatal(err)
	}
	if got := calls(db, model.NotificationCollection); len(got) != 1 {
		t.Errorf("calls = %v, want only the lookup", got)
	}

	if err := UpdateUserNotificationNotified(c, "nope", id.Hex()); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestDeleteUserNotification(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()

	newContext := func() (*mock.Database, *gin.Context) {
		db := mock.NewDatabase("app")
		db.Coll(model.NotificationCollection).One = mock.Result{Value: notificationDocument(id, userID, false)}
		db.Coll(model.NotificationCollection).Count = 4
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		return db, c
	}

	db, c := newContext()
	if err := DeleteUserNotification(c, userID.Hex(), id.Hex()); err != nil {
		t.Fatal(err)
	}

	// The counts of the user are refreshed.
	update := db.Coll(model.UserCollection).LastCall()
	if set := update.Args[1].(bson.M)["$set"].(bson.M); update.Args[0] != userID || set["notificationsStats.total"] != int64(4) {
		t.Errorf("update = %+v", update)
	}

	if err := DeleteUserNotification(c, "nope", id.Hex()); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	for _, method := range []string{"DeleteOne", "CountDocuments"} {
		db, c = newContext()
		db.Coll(model.NotificationCollection).Errs = map[string]error{method: errBoom}
		if err := DeleteUserNotification(c, userID.Hex(), id.Hex()); err == nil {
			t.Errorf("%s failing must be reported", method)
		}
	}

	// The count of the user is only logged.
	db, c = newContext()
	db.Coll(model.UserCollection).Err = errBoom
	if err := DeleteUserNotification(c, userID.Hex(), id.Hex()); err != nil {
		t.Errorf("err = %v", err)
	}
}

func TestSubscriptionToNotifications(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "notifications": []string{"news", "offers"}}}
	c, _ := plugintest.SignedInContext(t, db, model.User{ID: id})

	if err := SubscribeNotificationsToUser(c, id.Hex(), []string{"news"}); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$set": bson.M{"notifications": []string{"news"}}})

	if got := NotificationsSubscribedByUser(c, id.Hex()); len(got) != 2 || got[0] != "news" {
		t.Errorf("subscribed = %v", got)
	}

	// The user can not be loaded.
	c, _ = plugintest.SignedInContext(t, mock.NewDatabase("app"), model.User{ID: id})
	if got := NotificationsSubscribedByUser(c, id.Hex()); got == nil || len(got) != 0 {
		t.Errorf("subscribed = %v, want an empty list", got)
	}
}

func TestUnsubscribeNotificationFromUser(t *testing.T) {
	id := bson.NewObjectID()
	key := notifications.GenerateUnsubscribeKey(id.Hex(), "private", "news")

	newContext := func() (*mock.Database, *gin.Context) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "privateKey": "private"}}
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		return db, c
	}

	db, c := newContext()
	if err := UnsubscribeNotificationFromUser(c, id.Hex(), "news", key); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$pull": bson.M{"notifications": "news"}})

	// A key for another notification, another user or another private key is rejected.
	for _, wrong := range []string{"", "wrong", notifications.GenerateUnsubscribeKey(id.Hex(), "private", "offers"), notifications.GenerateUnsubscribeKey(id.Hex(), "other", "news")} {
		db, c = newContext()
		if err := UnsubscribeNotificationFromUser(c, id.Hex(), "news", wrong); err == nil {
			t.Errorf("the key %q must be rejected", wrong)
		}
		if got := calls(db, model.UserCollection); len(got) != 1 {
			t.Errorf("calls = %v: nothing must be updated", got)
		}
	}

	if err := UnsubscribeNotificationFromUser(c, "nope", "news", key); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	c, _ = plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
	if err := UnsubscribeNotificationFromUser(c, id.Hex(), "news", key); err == nil {
		t.Error("an unknown user must be reported")
	}

	db, c = newContext()
	db.Coll(model.UserCollection).Errs = map[string]error{"UpdateByID": errBoom}
	if err := UnsubscribeNotificationFromUser(c, id.Hex(), "news", key); err == nil {
		t.Error("a failure to update the user must be reported")
	}
}
