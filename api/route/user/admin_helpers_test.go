package user

import (
	"context"
	"errors"
	"net/http"
	"testing"

	commonErrors "github.com/cygy/ginamite/common/errors"

	"github.com/gin-gonic/gin"
)

func TestGetUserIDAndGetTokenID(t *testing.T) {
	c, _ := gin.CreateTestContext(nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	if got := getUserID(c); got != "abc" {
		t.Errorf("getUserID() = %q, want abc", got)
	}
	if got := getTokenID(c); got != "abc" {
		t.Errorf("getTokenID() = %q, want abc", got)
	}
}

func TestUpdateAbilities(t *testing.T) {
	handlers := []struct {
		name    string
		handler gin.HandlerFunc
		inject  func(fn func(ctx context.Context, userID string, abilities []string) error)
	}{
		{"SetAbilities", SetAbilities, func(fn func(context.Context, string, []string) error) { testStorage.SetAbilitiesToUserFn = fn }},
		{"AddAbilities", AddAbilities, func(fn func(context.Context, string, []string) error) { testStorage.AddAbilitiesToUserFn = fn }},
		{"RemoveAbilities", RemoveAbilities, func(fn func(context.Context, string, []string) error) { testStorage.RemoveAbilitiesFromUserFn = fn }},
	}
	tests := []struct {
		name    string
		body    string
		defined bool
		err     error
		want    int
	}{
		{"no abilities", `{"abilities":[]}`, true, nil, http.StatusBadRequest},
		{"storage not defined", `{"abilities":["a"]}`, false, nil, http.StatusInternalServerError},
		{"success", `{"abilities":["a"]}`, true, nil, http.StatusOK},
		{"user not found", `{"abilities":["a"]}`, true, commonErrors.NotFound(), http.StatusNotFound},
		{"other error", `{"abilities":["a"]}`, true, errors.New("boom"), http.StatusInternalServerError},
	}
	for _, h := range handlers {
		for _, tt := range tests {
			t.Run(h.name+"/"+tt.name, func(t *testing.T) {
				setup(t)

				var gotUser string
				var gotAbilities []string
				if !tt.defined {
					withoutStorage()
				}
				h.inject(func(_ context.Context, userID string, abilities []string) error {
					gotUser, gotAbilities = userID, abilities
					return tt.err
				})

				if got := serve(h.handler, tt.body); got != tt.want {
					t.Fatalf("status = %d, want %d", got, tt.want)
				}
				if tt.want == http.StatusOK && (gotUser != "42" || len(gotAbilities) != 1 || gotAbilities[0] != "a") {
					t.Errorf("function called with user %q and abilities %v", gotUser, gotAbilities)
				}
			})
		}
	}
}

func TestUpdateAbilities_UnknownAction(t *testing.T) {
	setup(t)

	c, recorder := newTestContext(`{"abilities":["a"]}`)
	updateAbilities(c, "unknown")
	c.Writer.WriteHeaderNow()

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}
