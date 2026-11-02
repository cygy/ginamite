package functions

import (
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestDoesAccountExist(t *testing.T) {
	tests := []struct {
		name string
		call func(db *mock.Database) (bool, error)
	}{
		{"username", func(db *mock.Database) (bool, error) {
			c, _ := plugintest.NewContext(db, "GET", "/", nil)
			return DoesAccountWithUsernameExist(c, "john")
		}},
		{"email address", func(db *mock.Database) (bool, error) {
			c, _ := plugintest.NewContext(db, "GET", "/", nil)
			return DoesAccountWithEmailAddressExist(c, "j@example.com")
		}},
		{"facebook ID", func(db *mock.Database) (bool, error) {
			c, _ := plugintest.NewContext(db, "GET", "/", nil)
			return DoesUserWithFacebookUserIDExist(c, "10157123")
		}},
		{"google ID", func(db *mock.Database) (bool, error) {
			c, _ := plugintest.NewContext(db, "GET", "/", nil)
			return DoesUserWithGoogleUserIDExist(c, "11800")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Nothing found.
			db := mock.NewDatabase("app")
			if got, err := tt.call(db); err != nil || got {
				t.Errorf("no account: got %v, %v", got, err)
			}
			if len(db.Coll(model.UserCollection).Calls()) != 1 || len(db.Coll(model.DisabledUserCollection).Calls()) != 1 {
				t.Error("both the enabled and the disabled users must be checked")
			}

			// Enabled user: the disabled users are not checked.
			db = mock.NewDatabase("app")
			db.Coll(model.UserCollection).Count = 1
			if got, err := tt.call(db); err != nil || !got {
				t.Errorf("enabled account: got %v, %v", got, err)
			}
			if len(db.Coll(model.DisabledUserCollection).Calls()) != 0 {
				t.Error("the disabled users must not be checked")
			}

			// Disabled user.
			db = mock.NewDatabase("app")
			db.Coll(model.DisabledUserCollection).Count = 1
			if got, err := tt.call(db); err != nil || !got {
				t.Errorf("disabled account: got %v, %v", got, err)
			}

			// Errors.
			for _, collection := range []string{model.UserCollection, model.DisabledUserCollection} {
				db = mock.NewDatabase("app")
				db.Coll(collection).Err = errBoom
				if got, err := tt.call(db); !errors.Is(err, errBoom) || got {
					t.Errorf("%s failing: got %v, %v", collection, got, err)
				}
			}
		})
	}
}
