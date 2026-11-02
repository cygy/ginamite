package functions

import (
	"testing"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestRegisterByEmailAddress(t *testing.T) {
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	err := RegisterByEmailAddress(c, "john", "hash", "j@example.com", "fr", "v2", "code", "key", "web", "1.2.3.4", authentication.Device{Name: "phone"})
	if err != nil {
		t.Fatal(err)
	}

	saved := db.Coll(model.UserCollection).LastCall().Args[0].(*model.User)
	if saved.Username != "john" || saved.IdentificationMethods.Password != "hash" || saved.RegistrationInfos.Code != "code" || saved.RegistrationInfos.Device.Name != "phone" {
		t.Errorf("user = %+v", saved)
	}

	db.Coll(model.UserCollection).Err = errBoom
	if err := RegisterByEmailAddress(c, "a", "b", "c", "fr", "v", "c", "k", "s", "ip", authentication.Device{}); err == nil {
		t.Error("a failure to save the user must be reported")
	}
}

func TestRegisterByThirdPartyToken(t *testing.T) {
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	infos := validator.TokenInfos{Token: "t", UserID: "123", Email: "j@example.com", FirstName: "John"}
	err := RegisterByThirdPartyToken(c, "john", infos, authentication.MethodGoogle, "fr", "v2", "code", "key", "web", "1.2.3.4", authentication.Device{})
	if err != nil {
		t.Fatal(err)
	}

	saved := db.Coll(model.UserCollection).LastCall().Args[0].(*model.User)
	if saved.IdentificationMethods.Google.UserID != "123" || !saved.Verified || saved.PrivateInfos.Email != "j@example.com" {
		t.Errorf("user = %+v", saved)
	}
}

func TestValidateUserRegistration(t *testing.T) {
	id := bson.NewObjectID()
	newContext := func(infos bson.M) (*mock.Database, func(key string) error) {
		db := mock.NewDatabase("app")
		doc := bson.M{"_id": id, "registrationInfos": bson.M{"code": "right"}}
		for k, v := range infos {
			doc[k] = v
		}
		db.Coll(model.UserCollection).One = mock.Result{Value: doc}
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		return db, func(key string) error { return ValidateUserRegistration(c, id.Hex(), key) }
	}

	t.Run("right key", func(t *testing.T) {
		db, validate := newContext(nil)
		if err := validate("right"); err != nil {
			t.Fatal(err)
		}
		if set := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M); set["verified"] != true {
			t.Errorf("set = %v", set)
		}
	})

	t.Run("wrong key", func(t *testing.T) {
		_, validate := newContext(nil)
		if err := validate("wrong"); err == nil {
			t.Error("a wrong key must be rejected")
		}
	})

	t.Run("already verified", func(t *testing.T) {
		db, validate := newContext(bson.M{"privateInfos": bson.M{"isEmailValid": true}})
		if err := validate("right"); err == nil {
			t.Error("a verified user must be rejected")
		}
		if got := calls(db, model.UserCollection); len(got) != 1 {
			t.Errorf("calls = %v, want only the lookup", got)
		}
	})

	t.Run("unknown user and invalid ID", func(t *testing.T) {
		c, _ := plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
		if err := ValidateUserRegistration(c, id.Hex(), "right"); err == nil {
			t.Error("an unknown user must be reported")
		}
		if err := ValidateUserRegistration(c, "nope", "right"); err == nil {
			t.Error("an invalid ID must be rejected")
		}
	})
}

func TestCancelUserRegistration(t *testing.T) {
	id := bson.NewObjectID()

	newContext := func(infos bson.M) (*mock.Database, func(key string) error) {
		db := mock.NewDatabase("app")
		doc := bson.M{"_id": id, "registrationInfos": bson.M{"code": "right"}}
		for k, v := range infos {
			doc[k] = v
		}
		db.Coll(model.UserCollection).One = mock.Result{Value: doc}
		c, _ := plugintest.NewContext(db, "GET", "/", nil)

		return db, func(key string) error { return CancelUserRegistration(c, id.Hex(), key) }
	}

	t.Run("right key", func(t *testing.T) {
		queue := plugintest.CaptureQueue(t)
		db, cancel := newContext(nil)

		if err := cancel("right"); err != nil {
			t.Fatal(err)
		}
		checkCall(t, db, model.UserCollection, "DeleteOne")

		// The workers delete the data of the user.
		if message, ok := queue.Next(); !ok || message.Type != "delete_user" {
			t.Errorf("message = %+v, %v", message, ok)
		}
	})

	t.Run("wrong key", func(t *testing.T) {
		db, cancel := newContext(nil)
		if err := cancel("wrong"); err == nil {
			t.Error("a wrong key must be rejected")
		}
		if got := calls(db, model.UserCollection); len(got) != 1 {
			t.Errorf("calls = %v, want only the lookup", got)
		}
	})

	t.Run("already verified", func(t *testing.T) {
		_, cancel := newContext(bson.M{"privateInfos": bson.M{"isEmailValid": true}})
		if err := cancel("right"); err == nil {
			t.Error("a verified user must be rejected")
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		c, _ := plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
		if err := CancelUserRegistration(c, "nope", "right"); err == nil {
			t.Error("an invalid ID must be rejected")
		}
	})
}

func TestGetRegistrationDetails(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "registrationInfos": bson.M{"source": "web", "method": "password", "device": bson.M{"name": "phone"}}}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	details, err := GetRegistrationDetails(c, id.Hex())
	if err != nil || details.Source != "web" || details.Method != "password" || details.Device.Name != "phone" {
		t.Errorf("got %+v, %v", details, err)
	}

	if _, err := GetRegistrationDetails(c, "nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}

	c, _ = plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
	if _, err := GetRegistrationDetails(c, id.Hex()); err == nil {
		t.Error("an unknown user must be reported")
	}
}
