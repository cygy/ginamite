package model

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestAuthToken_Document(t *testing.T) {
	testDocument(t, AuthTokenCollection, func() document.Document { return NewAuthToken() })
}

func TestNewAuthToken(t *testing.T) {
	if token := NewAuthToken(); token.CreatedAt.IsZero() {
		t.Error("the creation date must be set")
	}
}

func TestAuthToken_Updates(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()
	token := NewAuthToken()
	_ = token.SetID(hex)

	t.Run("UpdateProperties", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := token.UpdateProperties(ctx, "phone", true, []string{"a"}, db); err != nil {
			t.Fatal(err)
		}
		checkUpdate(t, db, AuthTokenCollection, id, bson.M{"$set": bson.M{
			"device.name":                               "phone",
			"device.notificationSettings.enabled":       true,
			"device.notificationSettings.notifications": []string{"a"},
		}})
	})

	t.Run("UpdateExpirationDate", func(t *testing.T) {
		db := mock.NewDatabase("app")
		date := time.Now().Add(time.Hour)
		if err := token.UpdateExpirationDate(ctx, date, db); err != nil {
			t.Fatal(err)
		}
		checkUpdate(t, db, AuthTokenCollection, id, bson.M{"$set": bson.M{"expiresAt": date}})
	})

	t.Run("UpdateLocation", func(t *testing.T) {
		db := mock.NewDatabase("app")
		location := IPLocation{IP: "1.2.3.4", City: "Paris"}
		if err := token.UpdateLocation(ctx, location, db); err != nil {
			t.Fatal(err)
		}

		set := lastCall(db, AuthTokenCollection).Args[1].(bson.M)["$set"].(bson.M)["location"].(IPLocation)
		if set.IP != "1.2.3.4" || set.City != "Paris" || !bson.IsObjectIDProvided(set.ID) {
			t.Errorf("location = %+v, a new ID is expected", set)
		}
		if bson.IsObjectIDProvided(location.ID) {
			t.Error("the location of the caller must not be modified")
		}
	})
}

func TestAuthToken_Token(t *testing.T) {
	userID, userHex := savedID()
	now := time.Now()

	token := &AuthToken{UserID: userID, Source: "web", Method: "password", Key: "k", CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	got := token.ToToken()
	if got.UserID != userHex || got.PrivateKey != "k" || got.Source != "web (password)" || !got.CreationDate.Equal(now) || !got.ExpirationDate.Equal(now.Add(time.Hour)) {
		t.Errorf("ToToken() = %+v", got)
	}

	var back AuthToken
	if err := back.FromToken(authentication.Token{UserID: userHex, Source: "web", PrivateKey: "k", CreationDate: now, ExpirationDate: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if back.UserID != userID || back.Source != "web" || back.Key != "k" || !back.CreatedAt.Equal(now) || !back.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Errorf("FromToken() = %+v", back)
	}

	if err := back.FromToken(authentication.Token{UserID: "nope"}); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestGetAuthTokens(t *testing.T) {
	ctx := context.Background()
	userID, userHex := savedID()

	db := mock.NewDatabase("app")
	db.Coll(AuthTokenCollection).Docs = []any{AuthToken{Source: "a", UserID: userID}, AuthToken{Source: "b", UserID: userID}}

	tokens, err := GetAuthTokens(ctx, userHex, db)
	if err != nil || len(tokens) != 2 || tokens[1].Source != "b" {
		t.Fatalf("got %+v, %v", tokens, err)
	}
	checkCall(t, db, AuthTokenCollection, "Find", bson.M{"userId": userID})
	if sort := mock.FindOptions(lastCall(db, AuthTokenCollection).Opts).Sort; sort == nil {
		t.Error("the tokens must be sorted")
	}

	if tokens, err := GetAuthTokens(ctx, "nope", db); err == nil || tokens == nil {
		t.Errorf("an invalid ID must fail with an empty list: %v, %v", tokens, err)
	}

	db.Coll(AuthTokenCollection).Err = errBoom
	if _, err := GetAuthTokens(ctx, userHex, db); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestUserFromAuthToken(t *testing.T) {
	_, hex := savedID()

	tests := []struct {
		name    string
		claims  jwt.Claims
		wantErr bool
		want    User
	}{
		{
			"all claims",
			jwt.MapClaims{"sub": hex, "nickname": "john", "locale": "fr-fr", "timezone": "Europe/Paris"},
			false,
			User{Username: "john", Settings: UserSettings{Locale: "fr-fr", Timezone: "Europe/Paris"}},
		},
		{"no claims", jwt.MapClaims{}, false, User{}},
		{"invalid subject", jwt.MapClaims{"sub": "nope"}, true, User{}},
		{"claims of the wrong type", jwt.MapClaims{"sub": 12, "nickname": 3, "locale": true, "timezone": []string{}}, false, User{}},
		{"not a map", jwt.RegisteredClaims{}, true, User{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := UserFromAuthToken(&jwt.Token{Claims: tt.claims})
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v", err)
			}
			if err != nil {
				return
			}

			if user.Username != tt.want.Username || user.Settings != tt.want.Settings {
				t.Errorf("user = %+v", user)
			}
			if _, isMap := tt.claims.(jwt.MapClaims); isMap && tt.claims.(jwt.MapClaims)["sub"] == hex && user.ID.Hex() != hex {
				t.Errorf("ID = %q", user.ID.Hex())
			}
		})
	}
}

func TestDeleteAuthTokens(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	t.Run("by ID", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := DeleteAuthTokenByID(ctx, hex, db); err != nil {
			t.Fatal(err)
		}
		checkCall(t, db, AuthTokenCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

		if err := DeleteAuthTokenByID(ctx, "nope", db); err == nil {
			t.Error("an invalid ID must be rejected")
		}

		db.Coll(AuthTokenCollection).DeleteResult = &mongo.DeleteResult{}
		if err := DeleteAuthTokenByID(ctx, hex, db); !errors.Is(err, mongo.ErrNoDocuments) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("by user", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(AuthTokenCollection).DeleteResult = &mongo.DeleteResult{DeletedCount: 3}

		n, err := DeleteAuthTokensByUser(ctx, hex, db)
		if err != nil || n != 3 {
			t.Errorf("got %d, %v", n, err)
		}
		checkCall(t, db, AuthTokenCollection, "DeleteMany", bson.M{"userId": id})

		if _, err := DeleteAuthTokensByUser(ctx, "nope", db); err == nil {
			t.Error("an invalid ID must be rejected")
		}
	})

	t.Run("expired", func(t *testing.T) {
		db := mock.NewDatabase("app")
		before := time.Now()

		if _, err := DeleteExpiredAuthTokens(ctx, db); err != nil {
			t.Fatal(err)
		}

		filter := lastCall(db, AuthTokenCollection).Args[0].(bson.M)["expiresAt"].(bson.M)["$lt"].(time.Time)
		if filter.Before(before) {
			t.Errorf("filter date = %v, want now", filter)
		}
	})
}

func TestIsValidAuthToken(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	tests := []struct {
		count int64
		want  bool
	}{{0, false}, {1, true}}
	for _, tt := range tests {
		db := mock.NewDatabase("app")
		db.Coll(AuthTokenCollection).Count = tt.count

		got, err := IsValidAuthToken(ctx, hex, db)
		if err != nil || got != tt.want {
			t.Errorf("count %d: got %v, %v", tt.count, got, err)
		}

		filter := lastCall(db, AuthTokenCollection).Args[0].(bson.D)
		if filter[0].Value != id || filter[1].Key != "expiresAt" {
			t.Errorf("filter = %v", filter)
		}
	}

	if _, err := IsValidAuthToken(ctx, "nope", mock.NewDatabase("app")); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}
