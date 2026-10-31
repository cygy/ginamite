package model

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/common/timezone"
)

func TestUser_RegisterByEmailAddress(t *testing.T) {
	ctx := context.Background()

	oldAbilities := DefaultUserAbilities
	defer func() { DefaultUserAbilities = oldAbilities }()
	DefaultUserAbilities = []string{"read"}

	db := mock.NewDatabase("app")
	user := NewUser()
	device := authentication.Device{Name: "phone", Type: "mobile"}

	err := user.RegisterByEmailAddress(ctx, "John Doe", "hash", "j@example.com", "fr-fr", "v2", "code", "key", "web", "1.2.3.4", device, "Europe/Paris", db)
	if err != nil {
		t.Fatal(err)
	}

	if user.Username != "John Doe" || user.Slug != "john-doe" || user.IdentificationMethods.Password != "hash" || user.PrivateInfos.Email != "j@example.com" || user.PrivateInfos.IsEmailValid ||
		user.Verified || user.VersionOfTermsAccepted != "v2" || user.PrivateKey != "key" {
		t.Errorf("user = %+v", user)
	}
	if user.Settings.Locale != "fr-fr" || user.Settings.Timezone != "Europe/Paris" {
		t.Errorf("settings = %+v", user.Settings)
	}
	reg := user.RegistrationInfos
	if reg.Code != "code" || reg.Source != "web" || reg.IPAddress != "1.2.3.4" || reg.Date.IsZero() || reg.Device.Name != "phone" {
		t.Errorf("registration = %+v", reg)
	}
	if !user.IsSaved() || lastCall(db, UserCollection).Method != "InsertOne" {
		t.Error("the user must be inserted")
	}

	// The default abilities are copied, not shared with the user.
	user.Abilities[0] = "root"
	if DefaultUserAbilities[0] != "read" {
		t.Error("changing the abilities of a user changed the default abilities")
	}

	db.Coll(UserCollection).Err = errBoom
	if err := NewUser().RegisterByEmailAddress(ctx, "a", "h", "e", "fr", "v", "c", "k", "s", "ip", device, "", db); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestRegistrationTimezone(t *testing.T) {
	timezones := &timezone.Configuration{TimezonesByCountries: map[string][]string{"FR": {"Europe/Paris", "Europe/Other"}}}

	tests := []struct {
		name            string
		timezones       *timezone.Configuration
		defaultTimezone string
		locale          string
		want            string
	}{
		{"timezone of the country", timezones, "UTC", "fr-FR", "Europe/Paris"},
		{"default timezone", &timezone.Configuration{TimezonesByCountries: map[string][]string{}}, "UTC", "fr-fr", "UTC"},
		{"no timezones configured", nil, "UTC", "fr-fr", "UTC"},
		{"no default timezone", nil, "", "fr-fr", ""},
		{"locale without country", timezones, "UTC", "fr", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RegistrationTimezone(tt.locale, tt.timezones, tt.defaultTimezone); got != tt.want {
				t.Errorf("RegistrationTimezone() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUser_ValidateRegistrationByEmailAddress(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	newUser := func(attempts uint) *User {
		user := NewUser()
		_ = user.SetID(hex)
		user.RegistrationInfos.Code = "right"
		user.RegistrationInfos.CodeAttempts = attempts

		return user
	}

	t.Run("right code", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := newUser(0).ValidateRegistrationByEmailAddress(ctx, "right", db); err != nil {
			t.Fatal(err)
		}

		update := lastCall(db, UserCollection).Args[1].(bson.M)
		set := update["$set"].(bson.M)
		if set["privateInfos.isEmailValid"] != true || set["verified"] != true || update["$unset"] == nil {
			t.Errorf("update = %v", update)
		}
	})

	t.Run("wrong code increments the attempts", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := newUser(0).ValidateRegistrationByEmailAddress(ctx, "wrong", db); err == nil {
			t.Fatal("expected an error")
		}
		checkUpdate(t, db, UserCollection, id, bson.M{"$inc": bson.M{"registrationInfos.codeAttempts": 1}})
	})

	t.Run("last allowed attempt increments the attempts", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := newUser(UserMaxValidateRegistrationAttemps-2).ValidateRegistrationByEmailAddress(ctx, "wrong", db); err == nil {
			t.Fatal("expected an error")
		}
		if update := lastCall(db, UserCollection).Args[1].(bson.M); update["$inc"] == nil {
			t.Errorf("update = %v", update)
		}
	})

	t.Run("too many attempts clear the registration", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := newUser(UserMaxValidateRegistrationAttemps-1).ValidateRegistrationByEmailAddress(ctx, "wrong", db); err == nil {
			t.Fatal("expected an error")
		}
		if update := lastCall(db, UserCollection).Args[1].(bson.M); update["$unset"] == nil || update["$inc"] != nil {
			t.Errorf("update = %v", update)
		}
	})

	t.Run("update failures are only logged for a wrong code", func(t *testing.T) {
		for _, attempts := range []uint{0, UserMaxValidateRegistrationAttemps - 1} {
			db := mock.NewDatabase("app")
			db.Coll(UserCollection).Err = errBoom

			err := newUser(attempts).ValidateRegistrationByEmailAddress(ctx, "wrong", db)
			if err == nil || errors.Is(err, errBoom) {
				t.Errorf("attempts %d: err = %v, want the wrong code error", attempts, err)
			}
		}
	})

	t.Run("right code, update failure", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(UserCollection).Err = errBoom
		if err := newUser(0).ValidateRegistrationByEmailAddress(ctx, "right", db); !errors.Is(err, errBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestUser_CancelRegistrationByEmailAddress(t *testing.T) {
	ctx := context.Background()
	_, hex := savedID()

	user := NewUser()
	_ = user.SetID(hex)
	user.RegistrationInfos.Code = "right"

	db := mock.NewDatabase("app")
	if err := user.CancelRegistrationByEmailAddress(ctx, "wrong", db); err == nil {
		t.Error("a wrong code must be rejected")
	}
	if calls := db.Coll(UserCollection).Calls(); len(calls) != 0 {
		t.Errorf("calls = %+v", calls)
	}

	if err := user.CancelRegistrationByEmailAddress(ctx, "right", db); err != nil {
		t.Fatal(err)
	}
	if lastCall(db, UserCollection).Method != "DeleteOne" {
		t.Error("the user must be deleted")
	}
}

func TestUser_RegisterByThirdPartyToken(t *testing.T) {
	ctx := context.Background()

	infos := validator.TokenInfos{Token: "t", UserID: "123", Email: "j@example.com", FirstName: "John", LastName: "Doe"}

	tests := []struct {
		method string
		oauth  func(u *User) UserOAuthInfos
	}{
		{authentication.MethodFacebook, func(u *User) UserOAuthInfos { return u.IdentificationMethods.Facebook }},
		{authentication.MethodGoogle, func(u *User) UserOAuthInfos { return u.IdentificationMethods.Google }},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			db := mock.NewDatabase("app")
			user := NewUser()

			if err := user.RegisterByThirdPartyToken(ctx, "john", infos, tt.method, "fr", "v2", "code", "key", "web", "1.2.3.4", authentication.Device{}, "", db); err != nil {
				t.Fatal(err)
			}

			if got := tt.oauth(user); got.UserID != "123" || got.Token != "t" {
				t.Errorf("oauth infos = %+v", got)
			}
			if user.PrivateInfos.FirstName != "John" || user.PrivateInfos.LastName != "Doe" || user.PrivateInfos.Email != "j@example.com" || !user.PrivateInfos.IsEmailValid || !user.Verified {
				t.Errorf("user = %+v", user)
			}
			if user.RegistrationInfos.Method != tt.method {
				t.Errorf("method = %q", user.RegistrationInfos.Method)
			}
			if !slices.Contains([]string{"InsertOne"}, lastCall(db, UserCollection).Method) {
				t.Error("the user must be inserted")
			}
		})
	}

	t.Run("unknown method", func(t *testing.T) {
		user := NewUser()
		if err := user.RegisterByThirdPartyToken(ctx, "john", infos, "twitter", "fr", "v", "c", "k", "s", "ip", authentication.Device{}, "", mock.NewDatabase("app")); err != nil {
			t.Fatal(err)
		}
		if user.IdentificationMethods.Facebook.UserID != "" || user.IdentificationMethods.Google.UserID != "" {
			t.Errorf("methods = %+v", user.IdentificationMethods)
		}
	})

	t.Run("error", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(UserCollection).Err = errBoom
		if err := NewUser().RegisterByThirdPartyToken(ctx, "john", infos, authentication.MethodGoogle, "fr", "v", "c", "k", "s", "ip", authentication.Device{}, "", db); !errors.Is(err, errBoom) {
			t.Errorf("err = %v", err)
		}
	})
}
