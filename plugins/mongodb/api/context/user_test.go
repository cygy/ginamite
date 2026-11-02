package context

import (
	"testing"

	ctxShared "github.com/cygy/ginamite/api/context/shared"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
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
			c, _ := plugintest.NewContext("GET", "/", nil)
			if tt.value != nil {
				c.Set(ctxShared.UserContextKey, tt.value)
			}

			if got := GetUser(c); got.ID != tt.want.ID || got.Username != tt.want.Username {
				t.Errorf("GetUser() = %+v", got)
			}
		})
	}
}
