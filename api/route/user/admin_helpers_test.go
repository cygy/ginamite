package user

import (
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
		inject  func(fn func(c *gin.Context, abilities []string, userID string) error)
	}{
		{"SetAbilities", SetAbilities, func(fn func(*gin.Context, []string, string) error) { SetAbilitiesToUser = fn }},
		{"AddAbilities", AddAbilities, func(fn func(*gin.Context, []string, string) error) { AddAbilitiesToUser = fn }},
		{"RemoveAbilities", RemoveAbilities, func(fn func(*gin.Context, []string, string) error) { RemoveAbilitiesFromUser = fn }},
	}
	tests := []struct {
		name    string
		body    string
		defined bool
		err     error
		want    int
	}{
		{"no abilities", `{"abilities":[]}`, true, nil, http.StatusBadRequest},
		{"function not defined", `{"abilities":["a"]}`, false, nil, http.StatusInternalServerError},
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
				if tt.defined {
					h.inject(func(_ *gin.Context, abilities []string, userID string) error {
						gotUser, gotAbilities = userID, abilities
						return tt.err
					})
				}

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
