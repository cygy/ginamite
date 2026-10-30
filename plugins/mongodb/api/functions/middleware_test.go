package functions

import (
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestIsValidAuthToken(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if ok, err := IsValidAuthToken(c, id.Hex()); err != nil || ok {
		t.Errorf("unknown token: got %v, %v", ok, err)
	}

	db.Coll(model.AuthTokenCollection).Count = 1
	if ok, err := IsValidAuthToken(c, id.Hex()); err != nil || !ok {
		t.Errorf("valid token: got %v, %v", ok, err)
	}
}

func TestGetUserAndLocaleFromAuthToken(t *testing.T) {
	id := bson.NewObjectID()

	token := &jwt.Token{Claims: jwt.MapClaims{"sub": id.Hex(), "nickname": "john", "locale": "fr-fr"}}
	user, locale, err := GetUserAndLocaleFromAuthToken(token)
	if err != nil || locale != "fr-fr" {
		t.Fatalf("got %v, %q, %v", user, locale, err)
	}

	// The user is stored by value: the context reads it back.
	got, ok := user.(model.User)
	if !ok || got.ID != id || got.Username != "john" {
		t.Errorf("user = %#v", user)
	}

	if _, _, err := GetUserAndLocaleFromAuthToken(&jwt.Token{Claims: jwt.MapClaims{"sub": "nope"}}); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestUserFromTheAuthTokenIsReadByTheContext(t *testing.T) {
	// The user built from the auth token must be readable by the other functions, with the real middleware.
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id}}

	c, _ := plugintest.SignedInContext(t, db, model.User{ID: id, Username: "john"})
	if err := UpdatedAccountInfos(c, id.Hex(), "John", "Doe"); err != nil {
		t.Fatal(err)
	}
}

func TestGetUserAbilities(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "isAdmin": true, "abilities": []string{"a"}, "versionOfTermsAccepted": "v2"}}
	c, _ := plugintest.SignedInContext(t, db, model.User{ID: id})

	isAdmin, abilities := GetUserAbilities(c, id.Hex())
	if !isAdmin || len(abilities) != 1 || abilities[0] != "a" {
		t.Errorf("got %v, %v", isAdmin, abilities)
	}
	if got := GetLatestVersionOfTermsAcceptedByUser(c, id.Hex()); got != "v2" {
		t.Errorf("terms = %q", got)
	}

	// The user can not be loaded.
	c, _ = plugintest.SignedInContext(t, mock.NewDatabase("app"), model.User{ID: id})
	if isAdmin, abilities := GetUserAbilities(c, id.Hex()); isAdmin || abilities != nil {
		t.Errorf("got %v, %v", isAdmin, abilities)
	}
	if got := GetLatestVersionOfTermsAcceptedByUser(c, id.Hex()); got != "" {
		t.Errorf("terms = %q", got)
	}
}
