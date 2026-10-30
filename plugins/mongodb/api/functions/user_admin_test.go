package functions

import (
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

func TestGetUsersForAdmin(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).Docs = []any{model.User{Username: "a"}}
	db.Coll(model.DisabledUserCollection).Docs = []any{model.DisabledUser{Username: "b"}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	enabled, err := GetEnabledUsersForAdmin(c)
	if users, ok := enabled.([]model.AdminUser); err != nil || !ok || len(users) != 1 || users[0].Username != "a" {
		t.Errorf("enabled users = %+v, %v", enabled, err)
	}

	disabled, err := GetDisabledUsersForAdmin(c)
	if users, ok := disabled.([]model.AdminDisabledUser); err != nil || !ok || len(users) != 1 || users[0].Username != "b" {
		t.Errorf("disabled users = %+v, %v", disabled, err)
	}
}

func TestUserUpdatesByAdmin(t *testing.T) {
	id := bson.NewObjectID()

	tests := []struct {
		name string
		call func(c *gin.Context, userID string) error
		want bson.M
	}{
		{"UpdateUserInformationsByAdmin", func(c *gin.Context, userID string) error {
			return UpdateUserInformationsByAdmin(c, userID, "John", "Doe", "j@example.com", true)
		}, bson.M{"$set": bson.M{"privateInfos.firstName": "John", "privateInfos.lastName": "Doe", "privateInfos.email": "j@example.com", "privateInfos.isEmailValid": true}}},
		{
			"UpdateUserDescriptionByAdmin", func(c *gin.Context, userID string) error {
				return UpdateUserDescriptionByAdmin(c, userID, "FR", "bonjour")
			},
			bson.M{"$set": bson.M{"publicInfos.description.fr": "bonjour"}},
		},
		{
			"UpdateUserSocialByAdmin", func(c *gin.Context, userID string) error { return UpdateUserSocialByAdmin(c, userID, "fb", "tw", "ig") },
			bson.M{"$set": bson.M{"publicInfos.facebook": "fb", "publicInfos.twitter": "tw", "publicInfos.instagram": "ig"}},
		},
		{
			"UpdateUserSettingsByAdmin", func(c *gin.Context, userID string) error {
				return UpdateUserSettingsByAdmin(c, userID, "fr-FR", "Europe/Paris")
			},
			bson.M{"$set": bson.M{"settings.locale": "fr-fr", "settings.timezone": "Europe/Paris"}},
		},
		{
			"UpdateUserNotificationsByAdmin", func(c *gin.Context, userID string) error {
				return UpdateUserNotificationsByAdmin(c, userID, []string{"a"})
			},
			bson.M{"$set": bson.M{"notifications": []string{"a"}}},
		},
		{
			"RemoveAbilitiesFromUser", func(c *gin.Context, userID string) error { return RemoveAbilitiesFromUser(c, []string{"a"}, userID) },
			bson.M{"$pullAll": bson.M{"abilities": []string{"a"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			c, _ := plugintest.NewContext(db, "GET", "/", nil)

			if err := tt.call(c, id.Hex()); err != nil {
				t.Fatal(err)
			}
			checkUpdate(t, db, model.UserCollection, id, tt.want)

			if err := tt.call(c, "nope"); err == nil {
				t.Error("an invalid user ID must be rejected")
			}
		})
	}
}

func TestUpdateUserPasswordByAdmin(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := UpdateUserPasswordByAdmin(c, id.Hex(), "Password1"); err != nil {
		t.Fatal(err)
	}
	stored := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["identificationMethods.password"]
	if stored == "Password1" {
		t.Error("the password must be encrypted")
	}

	if err := UpdateUserPasswordByAdmin(c, "nope", "Password1"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestAbilitiesOfUser(t *testing.T) {
	id := bson.NewObjectID()
	old := model.AllUserAbilities
	defer func() { model.AllUserAbilities = old }()
	model.AllUserAbilities = []string{"read", "write"}

	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := SetAbilitiesToUser(c, []string{"read", "root"}, id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$set": bson.M{"abilities": []string{"read"}}})

	if err := AddAbilitiesToUser(c, []string{"write", "root"}, id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$addToSet": bson.M{"abilities": bson.M{"$each": []string{"write"}}}})

	if SetAbilitiesToUser(c, nil, "nope") == nil || AddAbilitiesToUser(c, nil, "nope") == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestGetTokensForAdmin(t *testing.T) {
	userID := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.AuthTokenCollection).Docs = []any{model.AuthToken{Source: "web", UserID: userID}, model.AuthToken{Source: "mobile", UserID: userID}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	got, err := GetTokensForAdmin(c, userID.Hex())
	tokens, ok := got.([]*model.AdminAuthToken)
	if err != nil || !ok || len(tokens) != 2 || tokens[0].Source != "web" || tokens[1].Source != "mobile" {
		t.Errorf("got %+v, %v", got, err)
	}

	if _, err := GetTokensForAdmin(c, "nope"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestDeleteTokenByIDByAdmin(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := DeleteTokenByIDByAdmin(c, id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, model.AuthTokenCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})
}
