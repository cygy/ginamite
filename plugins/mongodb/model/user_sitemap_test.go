package model

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestNewSitemapUser(t *testing.T) {
	user := User{ID: mustObjectID(), Slug: "john", Username: "John"}
	if got := NewSitemapUser(user); got != (SitemapUser{ID: user.ID, Slug: "john"}) {
		t.Errorf("got %+v", got)
	}
}

func TestGetSitemapUsers(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll(UserCollection).Docs = []any{User{Slug: "a"}, User{Slug: "b"}}

	users, err := GetSitemapUsers(ctx, 5, 10, db)
	if err != nil || len(users) != 2 || users[1].Slug != "b" {
		t.Fatalf("got %+v, %v", users, err)
	}
	if opts := mock.FindOptions(lastCall(db, UserCollection).Opts); *opts.Skip != 5 || *opts.Limit != 10 || opts.Sort == nil {
		t.Errorf("options = %+v", opts)
	}

	db.Coll(UserCollection).Err = errBoom
	if users, err := GetSitemapUsers(ctx, 0, 1, db); !errors.Is(err, errBoom) || len(users) != 0 {
		t.Errorf("got %+v, %v", users, err)
	}
}
