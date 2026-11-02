package functions

import (
	"context"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

// tokenDocument returns an auth token of the user.
func tokenDocument(id, userID bson.ObjectID, expiresAt time.Time) bson.M {
	return bson.M{"_id": id, "userId": userID, "source": "web", "method": "password", "expiresAt": expiresAt, "createdAt": time.Now()}
}

func TestBuildTokenFromID(t *testing.T) {
	var tokens *Tokens
	id, userID := bson.NewObjectID(), bson.NewObjectID()
	expires := time.Now().Add(time.Hour).Truncate(time.Millisecond)

	db := mock.NewDatabase("app")
	db.Coll(model.AuthTokenCollection).One = mock.Result{Value: tokenDocument(id, userID, expires)}
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{
		"_id": userID, "username": "john", "settings": bson.M{"locale": "fr", "timezone": "Europe/Paris"},
		"publicInfos":        bson.M{"image": bson.M{"thumb": bson.M{"size1x": bson.M{"url": "thumb1"}, "size2x": bson.M{"url": "thumb2"}}}},
		"notificationsStats": bson.M{"unread": 3},
	}}
	tokens = tokensOn(db)

	token, err := tokens.BuildTokenFromID(context.Background(), id.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if token.UserID != userID.Hex() || token.Source != "web (password)" || !token.ExpirationDate.Equal(expires) {
		t.Errorf("token = %+v", token)
	}
	want := map[string]any{"nickname": "john", "locale": "fr", "timezone": "Europe/Paris", "img1": "thumb1", "img2": "thumb2", "unread": 3}
	for key, value := range want {
		if token.Other[key] != value {
			t.Errorf("%s = %v, want %v", key, token.Other[key], value)
		}
	}

	// Errors: the token, then the user, are not found.
	tokens = tokensOn(mock.NewDatabase("app"))
	if _, err := tokens.BuildTokenFromID(context.Background(), id.Hex()); err == nil {
		t.Error("an unknown token must be reported")
	}

	db = mock.NewDatabase("app")
	db.Coll(model.AuthTokenCollection).One = mock.Result{Value: tokenDocument(id, userID, expires)}
	tokens = tokensOn(db)
	if _, err := tokens.BuildTokenFromID(context.Background(), id.Hex()); err == nil {
		t.Error("an unknown user must be reported")
	}
}

func TestExtendTokenExpirationDateFromID(t *testing.T) {
	var tokens *Tokens
	id, userID := bson.NewObjectID(), bson.NewObjectID()
	const ttl = 3600 // in seconds

	tests := []struct {
		name       string
		expiresIn  time.Duration
		wantExtend bool
	}{
		{"less than half left", 10 * time.Minute, true},
		{"more than half left", 50 * time.Minute, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll(model.AuthTokenCollection).One = mock.Result{Value: tokenDocument(id, userID, time.Now().Add(tt.expiresIn))}
			tokens = tokensOn(db)

			if err := tokens.ExtendTokenExpirationDateFromID(context.Background(), id.Hex(), ttl); err != nil {
				t.Fatal(err)
			}

			updated := len(db.Coll(model.AuthTokenCollection).Calls()) == 2
			if updated != tt.wantExtend {
				t.Fatalf("updated = %v, calls = %v", updated, calls(db, model.AuthTokenCollection))
			}
			if tt.wantExtend {
				expiresAt := db.Coll(model.AuthTokenCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["expiresAt"].(time.Time)
				if time.Until(expiresAt) < 59*time.Minute {
					t.Errorf("the token expires in %v, want about an hour", time.Until(expiresAt))
				}
			}

			// The last login of the user is always updated.
			checkUpdate(t, db, model.UserCollection, userID, db.Coll(model.UserCollection).LastCall().Args[1].(bson.M))
			if _, ok := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)["lastLogin"]; !ok {
				t.Error("the last login must be updated")
			}
		})
	}

	t.Run("errors", func(t *testing.T) {
		tokens = tokensOn(mock.NewDatabase("app"))
		if err := tokens.ExtendTokenExpirationDateFromID(context.Background(), id.Hex(), ttl); err == nil {
			t.Error("an unknown token must be reported")
		}

		db := mock.NewDatabase("app")
		db.Coll(model.AuthTokenCollection).One = mock.Result{Value: tokenDocument(id, userID, time.Now())}
		db.Coll(model.AuthTokenCollection).Errs = map[string]error{"UpdateByID": errBoom}
		tokens = tokensOn(db)
		if err := tokens.ExtendTokenExpirationDateFromID(context.Background(), id.Hex(), ttl); err == nil {
			t.Error("a failure to extend the token must be reported")
		}

		// The last login is only logged.
		db = mock.NewDatabase("app")
		db.Coll(model.AuthTokenCollection).One = mock.Result{Value: tokenDocument(id, userID, time.Now().Add(time.Hour))}
		db.Coll(model.UserCollection).Err = errBoom
		tokens = tokensOn(db)
		if err := tokens.ExtendTokenExpirationDateFromID(context.Background(), id.Hex(), ttl); err != nil {
			t.Errorf("err = %v", err)
		}
	})
}

func TestTokens_SaveAuthenticationToken(t *testing.T) {
	userID := bson.NewObjectID()
	now := time.Now()
	details := authentication.Details{
		UserID: userID.Hex(), Source: "web", Method: "password", Key: "key", IPAddress: "1.2.3.4",
		CreationDate: now, ExpirationDate: now.Add(time.Hour), Device: authentication.Device{Name: "phone"},
	}

	t.Run("known IP address", func(t *testing.T) {
		db := mock.NewDatabase("app")
		locationID := bson.NewObjectID()
		db.Coll(model.IPLocationCollection).One = mock.Result{Value: bson.M{"_id": locationID, "IP": "1.2.3.4", "city": "Paris"}}
		tokens := tokensOn(db)

		tokenID, known, err := tokens.SaveAuthenticationToken(ctx, details)
		if err != nil || tokenID == "" || !known {
			t.Fatalf("got %q, %v, %v", tokenID, known, err)
		}

		saved := db.Coll(model.AuthTokenCollection).LastCall().Args[0].(*model.AuthToken)
		if saved.UserID != userID || saved.Key != "key" || saved.Device.Name != "phone" || saved.GetID() != tokenID || saved.Location == nil || saved.Location.City != "Paris" {
			t.Errorf("token = %+v", saved)
		}
		if saved.Location.ID == locationID {
			t.Error("the ID of the location document must not be saved in the token")
		}

		// The last login of the user is updated.
		if got := db.Coll(model.UserCollection).LastCall(); got.Method != "UpdateByID" || got.Args[0] != userID {
			t.Errorf("user call = %+v", got)
		}
	})

	t.Run("unknown IP address", func(t *testing.T) {
		db := mock.NewDatabase("app")
		tokens := tokensOn(db)

		_, known, err := tokens.SaveAuthenticationToken(ctx, details)
		if err != nil || known {
			t.Fatalf("got %v, %v", known, err)
		}
		if saved := db.Coll(model.AuthTokenCollection).LastCall().Args[0].(*model.AuthToken); saved.Location != nil {
			t.Errorf("location = %+v", saved.Location)
		}
	})

	t.Run("errors", func(t *testing.T) {
		tokens := tokensOn(mock.NewDatabase("app"))
		if _, _, err := tokens.SaveAuthenticationToken(ctx, authentication.Details{UserID: "nope"}); err == nil {
			t.Error("an invalid user ID must be rejected")
		}

		db := mock.NewDatabase("app")
		db.Coll(model.AuthTokenCollection).Err = errBoom
		tokens = tokensOn(db)
		if _, _, err := tokens.SaveAuthenticationToken(ctx, details); err == nil {
			t.Error("a failure to save the token must be reported")
		}

		// The last login is only logged.
		db = mock.NewDatabase("app")
		db.Coll(model.UserCollection).Err = errBoom
		tokens = tokensOn(db)
		if _, _, err := tokens.SaveAuthenticationToken(ctx, details); err != nil {
			t.Errorf("err = %v", err)
		}
	})
}

func TestTokens_Accessors(t *testing.T) {
	id, userID := bson.NewObjectID(), bson.NewObjectID()
	expires := time.Now().Add(time.Hour).Truncate(time.Millisecond)

	db := mock.NewDatabase("app")
	db.Coll(model.AuthTokenCollection).One = mock.Result{Value: tokenDocument(id, userID, expires)}
	tokens := tokensOn(db)

	if got, err := tokens.GetTokenExpirationDate(ctx, id.Hex()); err != nil || !got.Equal(expires) {
		t.Errorf("expiration date = %v, %v", got, err)
	}

	token, err := tokens.GetTokenDetailsByID(ctx, id.Hex())
	if err != nil || token.UserID != userID.Hex() || token.Source != "web (password)" {
		t.Errorf("details = %+v, %v", token, err)
	}

	own, ownerID, err := tokens.GetOwnedTokenByID(ctx, id.Hex())
	if ownToken, ok := own.(*model.OwnAuthToken); err != nil || !ok || ownToken.ID != id || ownerID != userID.Hex() {
		t.Errorf("got %+v, %q, %v", own, ownerID, err)
	}

	if err := tokens.DeleteTokenByID(ctx, id.Hex()); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, model.AuthTokenCollection, "DeleteOne", bson.D{{Key: "_id", Value: id}})

	// Unknown token.
	tokens = tokensOn(mock.NewDatabase("app"))
	if _, err := tokens.GetTokenExpirationDate(ctx, id.Hex()); err == nil {
		t.Error("an unknown token must be reported")
	}
	if _, err := tokens.GetTokenDetailsByID(ctx, id.Hex()); err == nil {
		t.Error("an unknown token must be reported")
	}
	if _, _, err := tokens.GetOwnedTokenByID(ctx, id.Hex()); err == nil {
		t.Error("an unknown token must be reported")
	}
}

func TestTokens_GetOwnedTokens(t *testing.T) {
	userID := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.AuthTokenCollection).Docs = []any{model.AuthToken{Source: "web"}, model.AuthToken{Source: "mobile"}}
	tokens := tokensOn(db)

	got, err := tokens.GetOwnedTokens(ctx, userID.Hex())
	owned, ok := got.([]*model.OwnAuthToken)
	if err != nil || !ok || len(owned) != 2 || owned[0].Source != "web" || owned[1].Source != "mobile" {
		t.Errorf("got %+v, %v", got, err)
	}

	if _, err := tokens.GetOwnedTokens(ctx, "nope"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	db.Coll(model.AuthTokenCollection).Err = errBoom
	if _, err := tokens.GetOwnedTokens(ctx, userID.Hex()); err == nil {
		t.Error("a failing query must be reported")
	}
}

func TestTokens_UpdateTokenByID(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	tokens := tokensOn(db)

	if err := tokens.UpdateTokenByID(ctx, id.Hex(), "phone", true, []string{"news"}); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.AuthTokenCollection, id, bson.M{"$set": bson.M{
		"device.name":                               "phone",
		"device.notificationSettings.enabled":       true,
		"device.notificationSettings.notifications": []string{"news"},
	}})

	if err := tokens.UpdateTokenByID(ctx, "nope", "phone", true, nil); err == nil {
		t.Error("an invalid token ID must be rejected")
	}
}
