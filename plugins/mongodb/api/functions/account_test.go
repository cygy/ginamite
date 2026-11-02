package functions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestGetOwnedAccountDetails(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "username": "john", "privateKey": "secret"}}

	got, err := accountsOn(db).GetOwnedAccountDetails(context.Background(), id.Hex())
	if user, ok := got.(*model.OwnUser); !ok || err != nil || user.ID != id || user.Username != "john" {
		t.Errorf("got %+v, %v", got, err)
	}

	if got, err := accountsOn(mock.NewDatabase("app")).GetOwnedAccountDetails(context.Background(), id.Hex()); err == nil || got != nil {
		t.Errorf("got %+v, %v, want an error when the user can not be loaded", got, err)
	}
}

func TestUpdatedAccountInfosAndSettings(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "username": "john"}}
	a := accountsOn(db)

	if err := a.UpdatedAccountInfos(context.Background(), id.Hex(), "John", "Doe"); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$set": bson.M{"privateInfos.firstName": "John", "privateInfos.lastName": "Doe"}})

	if err := a.UpdatedAccountSettings(context.Background(), id.Hex(), "FR", "Europe/Paris"); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$set": bson.M{"settings.locale": "fr", "settings.timezone": "Europe/Paris"}})

	if err := accountsOn(mock.NewDatabase("app")).UpdatedAccountInfos(context.Background(), id.Hex(), "John", "Doe"); err == nil {
		t.Error("an unknown user must be reported")
	}
}

// publicProfileRequest sends the request to update the public profile of the user.
func publicProfileRequest(t *testing.T, db *mock.Database, id bson.ObjectID, body string) *httptest.ResponseRecorder {
	t.Helper()

	c, recorder := plugintest.SignedInContext(t, model.User{ID: id})
	ctxkey.Config.Get(c).SupportedLocales = []string{"en", "fr"}
	c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer x")

	publicAccountOn(db, model.DefaultSettings()).UpdateProfile(c)

	return recorder
}

func TestPublicAccount_UpdateProfile(t *testing.T) {
	plugintest.InitLocalization(t, map[string]string{"status.account.public.profile.updated": "updated"})
	id := bson.NewObjectID()

	t.Run("updates the profile", func(t *testing.T) {
		queue := plugintest.CaptureQueue(t)
		db := mock.NewDatabase("app")

		recorder := publicProfileRequest(t, db, id, `{"locale":"FR","description":"bonjour"}`)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "updated") {
			t.Errorf("response = %d %s", recorder.Code, recorder.Body.String())
		}
		checkUpdate(t, db, model.UserCollection, id, bson.M{"$set": bson.M{"publicInfos.description.fr": "bonjour"}})

		// The workers update the documents that use the profile.
		if message, ok := queue.Next(); !ok || message.Type != "update_user_public_profile" {
			t.Errorf("message = %+v, %v", message, ok)
		}
	})

	t.Run("unsupported locale", func(t *testing.T) {
		db := mock.NewDatabase("app")

		recorder := publicProfileRequest(t, db, id, `{"locale":"de","description":"hallo"}`)
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status = %d", recorder.Code)
		}
		if got := calls(db, model.UserCollection); len(got) != 0 {
			t.Errorf("calls = %v", got)
		}
	})

	t.Run("description too long", func(t *testing.T) {
		db := mock.NewDatabase("app")

		body := `{"locale":"fr","description":"` + strings.Repeat("a", int(model.DefaultSettings().UserMaxDescriptionLength)+1) + `"}`
		if recorder := publicProfileRequest(t, db, id, body); recorder.Code != http.StatusBadRequest {
			t.Errorf("status = %d", recorder.Code)
		}
		if got := calls(db, model.UserCollection); len(got) != 0 {
			t.Errorf("calls = %v", got)
		}
	})

	t.Run("database error", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).Err = errBoom

		if recorder := publicProfileRequest(t, db, id, `{"locale":"fr","description":"bonjour"}`); recorder.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", recorder.Code)
		}
	})
}
