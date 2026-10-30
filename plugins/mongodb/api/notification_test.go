package api

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

var errBoom = errors.New("boom")

// withLimit sets the limit of notifications by user for the duration of the test.
func withLimit(t *testing.T, limit int64) {
	t.Helper()

	old := config.Main
	t.Cleanup(func() { config.Main = old })

	config.Main = config.NewConfiguration()
	config.Main.UserNotifications.Limit = limit
}

func TestCreateStoredNotification(t *testing.T) {
	ctx := context.Background()
	userID := bson.NewObjectID()
	locales := map[string]model.NotificationLocale{"en": {Title: "Title"}}
	image := model.PublicImage{Full: model.PublicImageSize{Size1x: "full1"}}

	t.Run("under the limit", func(t *testing.T) {
		withLimit(t, 10)
		db := mock.NewDatabase("app")
		db.Coll(model.NotificationCollection).Count = 3

		if err := CreateStoredNotification(ctx, userID.Hex(), "news", "data", image, locales, db); err != nil {
			t.Fatal(err)
		}

		notifications := db.Coll(model.NotificationCollection)
		insert := notifications.Calls()[0]
		saved := insert.Args[0].(*model.Notification)
		if insert.Method != "InsertOne" || saved.UserID != userID || saved.Type != "news" || saved.Data != "data" || saved.Locales["en"].Title != "Title" || saved.Image.Full.Size1x != "full1" {
			t.Errorf("notification = %+v", saved)
		}

		// The counts of the user are updated, nothing is deleted.
		update := db.Coll(model.UserCollection).LastCall()
		set := update.Args[1].(bson.M)["$set"].(bson.M)
		if update.Method != "UpdateByID" || update.Args[0] != userID || set["notificationsStats.total"] != int64(3) || set["notificationsStats.unread"] != int64(3) {
			t.Errorf("update = %+v", update)
		}
		for _, call := range notifications.Calls() {
			if call.Method == "DeleteMany" {
				t.Errorf("nothing must be deleted: %+v", call)
			}
		}
	})

	t.Run("image without URL is dropped", func(t *testing.T) {
		withLimit(t, 10)
		db := mock.NewDatabase("app")

		if err := CreateStoredNotification(ctx, userID.Hex(), "news", "", model.PublicImage{}, locales, db); err != nil {
			t.Fatal(err)
		}
		if saved := db.Coll(model.NotificationCollection).Calls()[0].Args[0].(*model.Notification); saved.Image.Full.Size1x != "" {
			t.Errorf("image = %+v", saved.Image)
		}
	})

	t.Run("over the limit", func(t *testing.T) {
		withLimit(t, 2)
		db := mock.NewDatabase("app")
		oldest := bson.NewObjectID()
		db.Coll(model.NotificationCollection).Count = 3
		db.Coll(model.NotificationCollection).Docs = []any{model.Notification{ID: oldest}}

		if err := CreateStoredNotification(ctx, userID.Hex(), "news", "", model.PublicImage{}, locales, db); err != nil {
			t.Fatal(err)
		}

		var deleted bool
		for _, call := range db.Coll(model.NotificationCollection).Calls() {
			if call.Method == "DeleteMany" {
				deleted = true
				if call.Args[0].(bson.M)["$or"].([]bson.M)[0]["_id"] != oldest {
					t.Errorf("filter = %v", call.Args[0])
				}
			}
		}
		if !deleted {
			t.Error("the extra notifications must be deleted")
		}
	})

	t.Run("failures to delete the extra notifications are only logged", func(t *testing.T) {
		withLimit(t, 2)

		for _, method := range []string{"Find", "DeleteMany"} {
			db := mock.NewDatabase("app")
			db.Coll(model.NotificationCollection).Count = 3
			db.Coll(model.NotificationCollection).Errs = map[string]error{method: errBoom}

			if err := CreateStoredNotification(ctx, userID.Hex(), "news", "", model.PublicImage{}, locales, db); err != nil {
				t.Errorf("%s failing: err = %v", method, err)
			}
		}
	})

	t.Run("errors", func(t *testing.T) {
		withLimit(t, 10)

		if err := CreateStoredNotification(ctx, "nope", "news", "", model.PublicImage{}, locales, mock.NewDatabase("app")); err == nil {
			t.Error("an invalid user ID must be rejected")
		}

		db := mock.NewDatabase("app")
		db.Coll(model.NotificationCollection).Err = errBoom
		if err := CreateStoredNotification(ctx, userID.Hex(), "news", "", model.PublicImage{}, locales, db); !errors.Is(err, errBoom) {
			t.Errorf("save err = %v", err)
		}

		// The notification is saved, the count fails.
		db = mock.NewDatabase("app")
		db.Coll(model.NotificationCollection).Errs = map[string]error{"CountDocuments": errBoom}
		if err := CreateStoredNotification(ctx, userID.Hex(), "news", "", model.PublicImage{}, locales, db); err == nil {
			t.Error("a count failure must be reported")
		}
	})

	t.Run("a failure to update the counts is only logged", func(t *testing.T) {
		withLimit(t, 10)
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).Err = errBoom

		if err := CreateStoredNotification(ctx, userID.Hex(), "news", "", model.PublicImage{}, locales, db); err != nil {
			t.Errorf("err = %v", err)
		}
	})
}
