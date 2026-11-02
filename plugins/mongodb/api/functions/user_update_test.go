package functions

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestUserUpdates(t *testing.T) {
	id := bson.NewObjectID()

	tests := []struct {
		name string
		call func(a *Accounts, c context.Context) error
		want bson.M
	}{
		{
			"UpdateUserEmailAddress", func(a *Accounts, c context.Context) error {
				return a.UpdateUserEmailAddress(c, id.Hex(), "new@example.com")
			},
			bson.M{"$set": bson.M{"privateInfos.email": "new@example.com", "privateInfos.isEmailValid": false}},
		},
		{
			"UserEmailAddressIsVerified", func(a *Accounts, c context.Context) error { return a.UserEmailAddressIsVerified(c, id.Hex()) },
			bson.M{"$set": bson.M{"privateInfos.isEmailValid": true}},
		},
		{
			"UpdateUserSocialNetworks", func(a *Accounts, c context.Context) error {
				return a.UpdateUserSocialNetworks(c, id.Hex(), "fb", "tw", "ig")
			},
			bson.M{"$set": bson.M{"publicInfos.facebook": "fb", "publicInfos.twitter": "tw", "publicInfos.instagram": "ig"}},
		},
		{
			"UpdateUserPassword already encrypted", func(a *Accounts, c context.Context) error { return a.UpdateUserPassword(c, id.Hex(), "hash", true) },
			bson.M{"$set": bson.M{"identificationMethods.password": "hash"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			c, a := context.Background(), accountsOn(db)

			if err := tt.call(a, c); err != nil {
				t.Fatal(err)
			}
			checkUpdate(t, db, model.UserCollection, id, tt.want)
		})
	}

	t.Run("invalid user ID", func(t *testing.T) {
		c, a := context.Background(), accountsOn(mock.NewDatabase("app"))

		for name, err := range map[string]error{
			"UpdateUserEmailAddress":     a.UpdateUserEmailAddress(c, "nope", "a"),
			"UserEmailAddressIsVerified": a.UserEmailAddressIsVerified(c, "nope"),
			"UpdateUserSocialNetworks":   a.UpdateUserSocialNetworks(c, "nope", "a", "b", "c"),
			"UpdateUserPassword":         a.UpdateUserPassword(c, "nope", "a", true),
		} {
			if err == nil {
				t.Errorf("%s: an invalid user ID must be rejected", name)
			}
		}
	})
}

func TestUpdateUserPassword_Encrypts(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, a := context.Background(), accountsOn(db)

	if err := a.UpdateUserPassword(c, id.Hex(), "Password1", false); err != nil {
		t.Fatal(err)
	}

	stored := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["identificationMethods.password"].(string)
	if !authentication.ComparePassword("Password1", stored) {
		t.Errorf("the password must be stored encrypted, got %q", stored)
	}
}
