package model

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestDisabledUser_Document(t *testing.T) {
	testDocument(t, DisabledUserCollection, func() document.Document { return &DisabledUser{} })
}

func TestNewDisabledUser(t *testing.T) {
	user := &User{
		ID: mustObjectID(), Username: "john", Slug: "john", Abilities: []string{"a"}, Notifications: []string{"n"},
		NotificationsStats:    UserNotificationsStats{Total: 3, Unread: 1},
		IdentificationMethods: UserIdentificationMethods{Password: "hash"},
		PublicInfos:           UserPublicInfos{Facebook: "fb"},
		PrivateInfos:          UserPrivateInfos{Email: "john@example.com"},
		RegistrationInfos:     UserRegistrationInfos{Source: "web"},
		Settings:              UserSettings{Locale: "fr"},
		IsAdmin:               true, IsAnonymous: true, Verified: true,
		VersionOfTermsAccepted: "v2", PrivateKey: "key",
	}

	disabled := NewDisabledUser(user)
	if disabled.UserID != user.ID || disabled.Username != "john" || disabled.PrivateKey != "key" || disabled.DisabledAt.IsZero() || !disabled.IsAdmin ||
		disabled.IdentificationMethods.Password != "hash" || disabled.NotificationsStats.Total != 3 || disabled.VersionOfTermsAccepted != "v2" {
		t.Errorf("got %+v", disabled)
	}
	if bson.IsObjectIDProvided(disabled.ID) {
		t.Error("the disabled user is a new document and needs its own ID")
	}

	// The user can be restored from the disabled user.
	restored := NewUserFromDisabledUser(disabled)
	if restored.ID != user.ID || restored.Username != "john" || restored.PrivateKey != "key" || !restored.IsAdmin || restored.Settings.Locale != "fr" ||
		restored.RegistrationInfos.Source != "web" || restored.PublicInfos.Facebook != "fb" {
		t.Errorf("restored = %+v", restored)
	}
}

func TestIsDisabledUserExisting(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	tests := []struct {
		name   string
		call   func(db *mock.Database) (bool, error)
		filter any
	}{
		{"username", func(db *mock.Database) (bool, error) { return IsDisabledUserWithUsernameExisting(ctx, "john", db) }, bson.M{"username": "john"}},
		{"email address", func(db *mock.Database) (bool, error) {
			return IsDisabledUserWithEmailAddressExisting(ctx, "john@example.com", db)
		}, bson.M{"privateInfos.email": "john@example.com"}},
		{"valid email address", func(db *mock.Database) (bool, error) {
			return IsDisabledUserWithValidEmailAddressExisting(ctx, "john@example.com", db)
		}, bson.M{"privateInfos.email": "john@example.com", "privateInfos.isEmailValid": true}},
		{
			"facebook ID", func(db *mock.Database) (bool, error) {
				return IsDisabledUserWithFacebookUserIDExisting(ctx, "123456", db)
			},
			bson.M{"identificationMethods.facebook.userId": "123456"},
		},
		{
			"google ID", func(db *mock.Database) (bool, error) { return IsDisabledUserWithGoogleUserIDExisting(ctx, hex, db) },
			bson.M{"identificationMethods.google.userId": hex},
		},
	}
	_ = id

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")

			if got, err := tt.call(db); err != nil || got {
				t.Errorf("without document: got %v, %v", got, err)
			}
			checkCall(t, db, DisabledUserCollection, "CountDocuments", tt.filter)

			db.Coll(DisabledUserCollection).Count = 1
			if got, err := tt.call(db); err != nil || !got {
				t.Errorf("with a document: got %v, %v", got, err)
			}

			db.Coll(DisabledUserCollection).Err = errBoom
			if _, err := tt.call(db); !errors.Is(err, errBoom) {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestDeleteDisabledUserByUserID(t *testing.T) {
	id, hex := savedID()

	db := mock.NewDatabase("app")
	if err := DeleteDisabledUserByUserID(context.Background(), hex, db); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, DisabledUserCollection, "DeleteMany", bson.M{"userId": id})

	if err := DeleteDisabledUserByUserID(context.Background(), "nope", db); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestGetDisabledUsersForAdmin(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll(DisabledUserCollection).Docs = []any{DisabledUser{Username: "a"}, DisabledUser{Username: "b", Verified: true}}

	users, err := GetDisabledUsersForAdmin(context.Background(), db)
	if err != nil || len(users) != 2 || users[1].Username != "b" || !users[1].Verified {
		t.Errorf("got %+v, %v", users, err)
	}
	if mock.FindOptions(lastCall(db, DisabledUserCollection).Opts).Sort == nil {
		t.Error("the users must be sorted")
	}
}
