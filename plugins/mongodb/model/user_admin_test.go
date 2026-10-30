package model

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestNewAdminUser(t *testing.T) {
	user := &User{ID: mustObjectID(), Username: "john", Verified: true, PrivateKey: "secret", PrivateInfos: UserPrivateInfos{Email: "j@example.com", IsEmailValid: true, FirstName: "John"}}

	got := NewAdminUser(user)
	want := AdminUser{ID: user.ID, Username: "john", Verified: true, PrivateInfos: AdminUserPrivateInfos{Email: "j@example.com", IsEmailValid: true}}
	if *got != want {
		t.Errorf("got %+v", got)
	}
}

func TestGetEnabledUsersForAdmin(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll(UserCollection).Docs = []any{User{Username: "a"}, User{Username: "b", Verified: true}}

	users, err := GetEnabledUsersForAdmin(context.Background(), db)
	if err != nil || len(users) != 2 || users[1].Username != "b" || !users[1].Verified {
		t.Errorf("got %+v, %v", users, err)
	}
	if mock.FindOptions(lastCall(db, UserCollection).Opts).Sort == nil {
		t.Error("the users must be sorted")
	}
}
