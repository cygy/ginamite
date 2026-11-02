package model

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestUser_Updates(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()
	user := NewUser()
	_ = user.SetID(hex)
	user.PublicInfos.Image = testImage()

	tests := []struct {
		name string
		call func(db *mock.Database) error
		want bson.M
	}{
		{"UpdateAllInformations", func(db *mock.Database) error {
			return user.UpdateAllInformations(ctx, "John", "Doe", "j@example.com", true, db)
		}, bson.M{"$set": bson.M{"privateInfos.firstName": "John", "privateInfos.lastName": "Doe", "privateInfos.email": "j@example.com", "privateInfos.isEmailValid": true}}},
		{
			"UpdateInfos", func(db *mock.Database) error { return user.UpdateInfos(ctx, "John", "Doe", db) },
			bson.M{"$set": bson.M{"privateInfos.firstName": "John", "privateInfos.lastName": "Doe"}},
		},
		{
			"UpdateSettings", func(db *mock.Database) error { return user.UpdateSettings(ctx, "FR-fr", "Europe/Paris", db) },
			bson.M{"$set": bson.M{"settings.locale": "fr-fr", "settings.timezone": "Europe/Paris"}},
		},
		{
			"UpdateEmailAddress", func(db *mock.Database) error { return user.UpdateEmailAddress(ctx, "new@example.com", db) },
			bson.M{"$set": bson.M{"privateInfos.email": "new@example.com", "privateInfos.isEmailValid": false}},
		},
		{
			"SetEmailAddressValid", func(db *mock.Database) error { return user.SetEmailAddressValid(ctx, db) },
			bson.M{"$set": bson.M{"privateInfos.isEmailValid": true}},
		},
		{
			"UpdatePublicProfile", func(db *mock.Database) error { return user.UpdatePublicProfile(ctx, "FR", "bonjour", db) },
			bson.M{"$set": bson.M{"publicInfos.description.fr": "bonjour"}},
		},
		{
			"UpdateSocialNetworks", func(db *mock.Database) error { return user.UpdateSocialNetworks(ctx, "fb", "tw", "ig", db) },
			bson.M{"$set": bson.M{"publicInfos.facebook": "fb", "publicInfos.twitter": "tw", "publicInfos.instagram": "ig"}},
		},
		{
			"AcceptVersionOfTerms", func(db *mock.Database) error { return user.AcceptVersionOfTerms(ctx, "v3", db) },
			bson.M{"$set": bson.M{"versionOfTermsAccepted": "v3"}},
		},
		{
			"UpdatePassword already encrypted", func(db *mock.Database) error { return user.UpdatePassword(ctx, "hash", true, db) },
			bson.M{"$set": bson.M{"identificationMethods.password": "hash"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			if err := tt.call(db); err != nil {
				t.Fatal(err)
			}
			checkUpdate(t, db, UserCollection, id, tt.want)
		})
	}
}

func TestUser_UpdatePassword(t *testing.T) {
	_, hex := savedID()
	user := NewUser()
	_ = user.SetID(hex)

	db := mock.NewDatabase("app")
	if err := user.UpdatePassword(context.Background(), "Password1", false, db); err != nil {
		t.Fatal(err)
	}

	stored := lastCall(db, UserCollection).Args[1].(bson.M)["$set"].(bson.M)["identificationMethods.password"].(string)
	if stored == "Password1" || !authentication.ComparePassword("Password1", stored) {
		t.Errorf("the password must be stored encrypted, got %q", stored)
	}
}

func TestUser_UpdateImage(t *testing.T) {
	id, hex := savedID()
	user := NewUser()
	_ = user.SetID(hex)
	user.PublicInfos.Image = testImage()
	user.PublicInfos.Image.Full.Size1x.Path = "/images/full1.png"

	db := mock.NewDatabase("app")
	if err := user.UpdateImage(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	call := lastCall(db, UserCollection)
	if call.Args[0] != id {
		t.Fatalf("call = %+v", call)
	}

	set := call.Args[1].(bson.M)["$set"].(bson.M)
	if len(set) != 12 || set["publicInfos.image.full.size1x.path"] != "/images/full1.png" || set["publicInfos.image.full.size1x.url"] != "full1" ||
		set["publicInfos.image.small.size2x.url"] != "small2" || set["publicInfos.image.thumb.size1x.url"] != "thumb1" {
		t.Errorf("set = %v", set)
	}
}
