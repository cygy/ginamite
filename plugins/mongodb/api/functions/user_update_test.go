package functions

import (
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestUserUpdates(t *testing.T) {
	id := bson.NewObjectID()

	tests := []struct {
		name string
		call func(c *gin.Context) error
		want bson.M
	}{
		{
			"UpdateUserEmailAddress", func(c *gin.Context) error { return UpdateUserEmailAddress(c, id.Hex(), "new@example.com") },
			bson.M{"$set": bson.M{"privateInfos.email": "new@example.com", "privateInfos.isEmailValid": false}},
		},
		{
			"UserEmailAddressIsVerified", func(c *gin.Context) error { return UserEmailAddressIsVerified(c, id.Hex()) },
			bson.M{"$set": bson.M{"privateInfos.isEmailValid": true}},
		},
		{
			"UpdateUserSocialNetworks", func(c *gin.Context) error { return UpdateUserSocialNetworks(c, id.Hex(), "fb", "tw", "ig") },
			bson.M{"$set": bson.M{"publicInfos.facebook": "fb", "publicInfos.twitter": "tw", "publicInfos.instagram": "ig"}},
		},
		{
			"UpdateUserPassword already encrypted", func(c *gin.Context) error { return UpdateUserPassword(c, id.Hex(), "hash", true) },
			bson.M{"$set": bson.M{"identificationMethods.password": "hash"}},
		},
		{
			"UpdateVersionOfTermsAcceptedByUser", func(c *gin.Context) error { return UpdateVersionOfTermsAcceptedByUser(c, id.Hex(), "v3") },
			bson.M{"$set": bson.M{"versionOfTermsAccepted": "v3"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			c, _ := plugintest.SignedInContext(t, db, model.User{ID: id})

			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			checkUpdate(t, db, model.UserCollection, id, tt.want)
		})
	}

	t.Run("invalid user ID", func(t *testing.T) {
		c, _ := plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)

		for name, err := range map[string]error{
			"UpdateUserEmailAddress":     UpdateUserEmailAddress(c, "nope", "a"),
			"UserEmailAddressIsVerified": UserEmailAddressIsVerified(c, "nope"),
			"UpdateUserSocialNetworks":   UpdateUserSocialNetworks(c, "nope", "a", "b", "c"),
			"UpdateUserPassword":         UpdateUserPassword(c, "nope", "a", true),
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
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := UpdateUserPassword(c, id.Hex(), "Password1", false); err != nil {
		t.Fatal(err)
	}

	stored := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["identificationMethods.password"].(string)
	if !authentication.ComparePassword("Password1", stored) {
		t.Errorf("the password must be stored encrypted, got %q", stored)
	}
}
