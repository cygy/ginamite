package functions

import (
	"context"
	"testing"
	"time"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestSaveUserTokenDetailsToUser(t *testing.T) {
	id := bson.NewObjectID()
	infos := validator.TokenInfos{Token: "t", UserID: "123", Email: "j@example.com", FirstName: "John", LastName: "Doe", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}

	tests := []struct {
		name   string
		call   func(a *Accounts, c context.Context, userID string) error
		method string
	}{
		{"facebook", func(a *Accounts, c context.Context, userID string) error {
			return a.SaveFacebookUserTokenDetailsToUser(c, userID, infos)
		}, authentication.MethodFacebook},
		{"google", func(a *Accounts, c context.Context, userID string) error {
			return a.SaveGoogleUserTokenDetailsToUser(c, userID, infos)
		}, authentication.MethodGoogle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id}}
			c, a := context.Background(), accountsOn(db)

			if err := tt.call(a, c, id.Hex()); err != nil {
				t.Fatal(err)
			}

			set := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)
			if saved, ok := set["identificationMethods."+tt.method].(model.UserOAuthInfos); !ok || saved.UserID != "123" || saved.Token != "t" {
				t.Errorf("set = %v", set)
			}

			// The invalid IDs and the unknown users are rejected before any update.
			if err := tt.call(a, c, "nope"); err == nil {
				t.Error("an invalid user ID must be rejected")
			}

			db = mock.NewDatabase("app")
			a = accountsOn(db)
			db.Coll(model.UserCollection).Err = errBoom
			if err := tt.call(a, c, id.Hex()); err == nil {
				t.Error("a failing lookup must be reported")
			}
		})
	}
}
