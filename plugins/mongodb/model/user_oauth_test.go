package model

import (
	"context"
	"testing"
	"time"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestUser_SaveThirdPartyInfos(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()
	created := time.Now()
	expires := created.Add(time.Hour)

	tests := []struct {
		name   string
		call   func(user *User, db *mock.Database) error
		method string
	}{
		{"facebook", func(u *User, db *mock.Database) error {
			return u.SaveFacebookInfos(ctx, "token", "123", "John", "Doe", "j@example.com", created, expires, db)
		}, authentication.MethodFacebook},
		{"google", func(u *User, db *mock.Database) error {
			return u.SaveGoogleInfos(ctx, "token", "123", "John", "Doe", "j@example.com", created, expires, db)
		}, authentication.MethodGoogle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser()
			_ = user.SetID(hex)
			db := mock.NewDatabase("app")

			if err := tt.call(user, db); err != nil {
				t.Fatal(err)
			}

			set := lastCall(db, UserCollection).Args[1].(bson.M)["$set"].(bson.M)
			infos, ok := set["identificationMethods."+tt.method].(UserOAuthInfos)
			if !ok || infos.UserID != "123" || infos.Token != "token" || infos.Email != "j@example.com" || !infos.ExpiresAt.Equal(expires) {
				t.Errorf("set = %v", set)
			}
			// The user has no name and a different email address: the names are filled, the email address is not validated.
			if set["privateInfos.firstName"] != "John" || set["privateInfos.lastName"] != "Doe" {
				t.Errorf("the names must be filled: %v", set)
			}
			if _, ok := set["privateInfos.isEmailValid"]; ok {
				t.Errorf("the email address of the user is not the one of the network: %v", set)
			}
			if lastCall(db, UserCollection).Args[0] != id {
				t.Error("the wrong document is updated")
			}
		})
	}
}

func TestUser_SaveThirdPartyInfos_Existing(t *testing.T) {
	ctx := context.Background()
	_, hex := savedID()

	user := NewUser()
	_ = user.SetID(hex)
	user.PrivateInfos = UserPrivateInfos{FirstName: "Jo", LastName: "Do", Email: "J@Example.com"}

	db := mock.NewDatabase("app")
	if err := user.SaveGoogleInfos(ctx, "t", "1", "John", "Doe", "j@example.com", time.Now(), time.Now(), db); err != nil {
		t.Fatal(err)
	}

	set := lastCall(db, UserCollection).Args[1].(bson.M)["$set"].(bson.M)
	if _, ok := set["privateInfos.firstName"]; ok {
		t.Errorf("the existing names must be kept: %v", set)
	}
	if set["privateInfos.isEmailValid"] != true {
		t.Errorf("the email address, validated by the network, must be validated here too: %v", set)
	}

	// Nothing to validate without email address.
	user.PrivateInfos.IsEmailValid = false
	if err := user.SaveGoogleInfos(ctx, "t", "1", "", "", "", time.Now(), time.Now(), db); err != nil {
		t.Fatal(err)
	}
	set = lastCall(db, UserCollection).Args[1].(bson.M)["$set"].(bson.M)
	if _, ok := set["privateInfos.isEmailValid"]; ok || len(set) != 1 {
		t.Errorf("set = %v", set)
	}
}

func TestUserOAuthInfos_FromSocialNetworkOAuthInfos(t *testing.T) {
	now := time.Now()
	infos := UserOAuthInfos{}
	infos.FromSocialNetworkOAuthInfos(validator.TokenInfos{Token: "t", UserID: "1", Email: "j@example.com", FirstName: "John", LastName: "Doe", CreatedAt: now, ExpiresAt: now.Add(time.Hour)})

	want := UserOAuthInfos{UserID: "1", Token: "t", FirstName: "John", LastName: "Doe", Email: "j@example.com", CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	if infos != want {
		t.Errorf("got %+v", infos)
	}
}
