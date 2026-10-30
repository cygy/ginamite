package functions

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

func TestSaveUserTokenDetailsToUser(t *testing.T) {
	id := bson.NewObjectID()
	infos := validator.TokenInfos{Token: "t", UserID: "123", Email: "j@example.com", FirstName: "John", LastName: "Doe", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}

	tests := []struct {
		name   string
		call   func(c *gin.Context, userID string) error
		method string
	}{
		{"facebook", func(c *gin.Context, userID string) error { return SaveFacebookUserTokenDetailsToUser(c, userID, infos) }, authentication.MethodFacebook},
		{"google", func(c *gin.Context, userID string) error { return SaveGoogleUserTokenDetailsToUser(c, userID, infos) }, authentication.MethodGoogle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id}}
			c, _ := plugintest.NewContext(db, "GET", "/", nil)

			if err := tt.call(c, id.Hex()); err != nil {
				t.Fatal(err)
			}

			set := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)
			if saved, ok := set["identificationMethods."+tt.method].(model.UserOAuthInfos); !ok || saved.UserID != "123" || saved.Token != "t" {
				t.Errorf("set = %v", set)
			}

			// The invalid IDs and the unknown users are rejected before any update.
			if err := tt.call(c, "nope"); err == nil {
				t.Error("an invalid user ID must be rejected")
			}

			db = mock.NewDatabase("app")
			c, _ = plugintest.NewContext(db, "GET", "/", nil)
			db.Coll(model.UserCollection).Err = errBoom
			if err := tt.call(c, id.Hex()); err == nil {
				t.Error("a failing lookup must be reported")
			}
		})
	}
}
