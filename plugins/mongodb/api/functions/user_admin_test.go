package functions

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

// newUsers returns the storage of the users working on the database, with the default settings.
func newUsers(db *mock.Database) *Users {
	return NewUsers(func() store.Database { return db }, model.DefaultSettings)
}

func TestUsers_GetUsersForAdmin(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).Docs = []any{model.User{Username: "a"}}
	db.Coll(model.DisabledUserCollection).Docs = []any{model.DisabledUser{Username: "b"}}
	users := newUsers(db)

	enabled, err := users.GetEnabledUsersForAdmin(ctx)
	if got, ok := enabled.([]model.AdminUser); err != nil || !ok || len(got) != 1 || got[0].Username != "a" {
		t.Errorf("enabled users = %+v, %v", enabled, err)
	}

	disabled, err := users.GetDisabledUsersForAdmin(ctx)
	if got, ok := disabled.([]model.AdminDisabledUser); err != nil || !ok || len(got) != 1 || got[0].Username != "b" {
		t.Errorf("disabled users = %+v, %v", disabled, err)
	}
}

func TestUsers_UpdatesByAdmin(t *testing.T) {
	id := bson.NewObjectID()

	tests := []struct {
		name string
		call func(users *Users, userID string) error
		want bson.M
	}{
		{"UpdateInformationsByAdmin", func(users *Users, userID string) error {
			return users.UpdateInformationsByAdmin(ctx, userID, "John", "Doe", "j@example.com", true)
		}, bson.M{"$set": bson.M{"privateInfos.firstName": "John", "privateInfos.lastName": "Doe", "privateInfos.email": "j@example.com", "privateInfos.isEmailValid": true}}},
		{
			"UpdateDescriptionByAdmin", func(users *Users, userID string) error {
				return users.UpdateDescriptionByAdmin(ctx, userID, "FR", "bonjour")
			},
			bson.M{"$set": bson.M{"publicInfos.description.fr": "bonjour"}},
		},
		{
			"UpdateSocialByAdmin", func(users *Users, userID string) error {
				return users.UpdateSocialByAdmin(ctx, userID, "fb", "tw", "ig")
			},
			bson.M{"$set": bson.M{"publicInfos.facebook": "fb", "publicInfos.twitter": "tw", "publicInfos.instagram": "ig"}},
		},
		{
			"UpdateSettingsByAdmin", func(users *Users, userID string) error {
				return users.UpdateSettingsByAdmin(ctx, userID, "fr-FR", "Europe/Paris")
			},
			bson.M{"$set": bson.M{"settings.locale": "fr-fr", "settings.timezone": "Europe/Paris"}},
		},
		{
			"UpdateNotificationsByAdmin", func(users *Users, userID string) error {
				return users.UpdateNotificationsByAdmin(ctx, userID, []string{"a"})
			},
			bson.M{"$set": bson.M{"notifications": []string{"a"}}},
		},
		{
			"RemoveAbilitiesFromUser", func(users *Users, userID string) error {
				return users.RemoveAbilitiesFromUser(ctx, userID, []string{"a"})
			},
			bson.M{"$pullAll": bson.M{"abilities": []string{"a"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			users := newUsers(db)

			if err := tt.call(users, id.Hex()); err != nil {
				t.Fatal(err)
			}
			checkUpdate(t, db, model.UserCollection, id, tt.want)

			if err := tt.call(users, "nope"); err == nil {
				t.Error("an invalid user ID must be rejected")
			}
		})
	}
}

func TestUsers_UpdatePasswordByAdmin(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	users := newUsers(db)

	if err := users.UpdatePasswordByAdmin(ctx, id.Hex(), "Password1"); err != nil {
		t.Fatal(err)
	}
	stored := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["identificationMethods.password"]
	if stored == "Password1" {
		t.Error("the password must be encrypted")
	}

	if err := users.UpdatePasswordByAdmin(ctx, "nope", "Password1"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestUsers_Abilities(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	settings := model.DefaultSettings()
	settings.AllUserAbilities = []string{"read", "write"}
	users := NewUsers(func() store.Database { return db }, func() model.Settings { return settings })

	if err := users.SetAbilitiesToUser(ctx, id.Hex(), []string{"read", "root"}); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$set": bson.M{"abilities": []string{"read"}}})

	if err := users.AddAbilitiesToUser(ctx, id.Hex(), []string{"write", "root"}); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$addToSet": bson.M{"abilities": bson.M{"$each": []string{"write"}}}})

	if users.SetAbilitiesToUser(ctx, "nope", nil) == nil || users.AddAbilitiesToUser(ctx, "nope", nil) == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestUsers_GetTokensForAdmin(t *testing.T) {
	userID := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.AuthTokenCollection).Docs = []any{model.AuthToken{Source: "web", UserID: userID}, model.AuthToken{Source: "mobile", UserID: userID}}
	users := newUsers(db)

	got, err := users.GetTokensForAdmin(ctx, userID.Hex())
	tokens, ok := got.([]*model.AdminAuthToken)
	if err != nil || !ok || len(tokens) != 2 || tokens[0].Source != "web" || tokens[1].Source != "mobile" {
		t.Errorf("got %+v, %v", got, err)
	}

	if _, err := users.GetTokensForAdmin(ctx, "nope"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestUsers_DeleteTokenByIDByAdmin(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")

	if err := newUsers(db).DeleteTokenByIDByAdmin(ctx, id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, model.AuthTokenCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})
}

func TestUsers_UserAbilities(t *testing.T) {
	ctx := context.Background()
	id := bson.NewObjectID()

	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "isAdmin": true, "abilities": []string{"a"}, "versionOfTermsAccepted": "v2"}}
	isAdmin, abilities, err := newUsers(db).Abilities(ctx, id.Hex())
	if err != nil || !isAdmin || len(abilities) != 1 || abilities[0] != "a" {
		t.Errorf("got %v, %v, %v", isAdmin, abilities, err)
	}

	// An unknown user has no right.
	if isAdmin, abilities, err := newUsers(mock.NewDatabase("app")).Abilities(ctx, id.Hex()); err != nil || isAdmin || abilities != nil {
		t.Errorf("unknown user: got %v, %v, %v", isAdmin, abilities, err)
	}

	// The failure of the database is reported.
	failing := mock.NewDatabase("app")
	failing.Coll(model.UserCollection).Err = errBoom
	if _, _, err := newUsers(failing).Abilities(ctx, id.Hex()); !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want the error of the database", err)
	}
}
