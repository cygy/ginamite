package context

import (
	"testing"

	ctxShared "github.com/cygy/ginamite/api/context/shared"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestGetUser(t *testing.T) {
	user := model.User{ID: bson.NewObjectID(), Username: "john"}

	tests := []struct {
		name  string
		value any
		want  model.User
	}{
		{"value", user, user},
		{"pointer", &user, user},
		{"nil pointer", (*model.User)(nil), model.User{}},
		{"unexpected type", "john", model.User{}},
		{"no user", nil, model.User{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
			if tt.value != nil {
				c.Set(ctxShared.UserContextKey, tt.value)
			}

			if got := GetUser(c); got.ID != tt.want.ID || got.Username != tt.want.Username {
				t.Errorf("GetUser() = %+v", got)
			}
		})
	}
}

func TestGetFullUser(t *testing.T) {
	id := bson.NewObjectID()

	t.Run("loaded once", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "username": "john"}}
		c, _ := plugintest.SignedInContext(t, db, model.User{ID: id})

		user := GetFullUser(c)
		if user == nil || user.Username != "john" {
			t.Fatalf("user = %+v", user)
		}
		if got := GetFullUser(c); got != user {
			t.Error("the user must be cached in the context")
		}
		if n := len(db.Coll(model.UserCollection).Calls()); n != 1 {
			t.Errorf("%d queries, want 1", n)
		}
	})

	t.Run("not found", func(t *testing.T) {
		db := mock.NewDatabase("app")
		c, _ := plugintest.SignedInContext(t, db, model.User{ID: id})

		if user := GetFullUser(c); user != nil {
			t.Errorf("user = %+v", user)
		}
	})

	t.Run("database error", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).Err = mongo.ErrClientDisconnected
		c, _ := plugintest.SignedInContext(t, db, model.User{ID: id})

		if user := GetFullUser(c); user != nil {
			t.Errorf("user = %+v", user)
		}
	})

	t.Run("not signed in", func(t *testing.T) {
		c, _ := plugintest.NewContext(mock.NewDatabase("app"), "GET", "/", nil)
		if user := GetFullUser(c); user != nil {
			t.Errorf("user = %+v", user)
		}
	})
}
