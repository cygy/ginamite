package functions

import (
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"

	jwt "github.com/golang-jwt/jwt/v5"
)

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
