package model

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestUser_HasAbility(t *testing.T) {
	user := &User{Abilities: []string{"a", "b"}}

	if !user.HasAbility("b") || user.HasAbility("c") || (&User{}).HasAbility("a") {
		t.Error("HasAbility() does not match the abilities of the user")
	}
}

func TestUser_Abilities(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()
	user := NewUser()
	_ = user.SetID(hex)

	allAbilities := []string{"read", "write"}

	// The unknown abilities are ignored.
	t.Run("SetAbilities", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := user.SetAbilities(ctx, []string{"write", "root", "read"}, allAbilities, db); err != nil {
			t.Fatal(err)
		}
		checkUpdate(t, db, UserCollection, id, bson.M{"$set": bson.M{"abilities": []string{"write", "read"}}})
	})

	t.Run("AddAbilities", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := user.AddAbilities(ctx, []string{"root", "read"}, allAbilities, db); err != nil {
			t.Fatal(err)
		}
		checkUpdate(t, db, UserCollection, id, bson.M{"$addToSet": bson.M{"abilities": bson.M{"$each": []string{"read"}}}})
	})

	t.Run("RemoveAbilities", func(t *testing.T) {
		db := mock.NewDatabase("app")
		if err := user.RemoveAbilities(ctx, []string{"read"}, db); err != nil {
			t.Fatal(err)
		}
		checkUpdate(t, db, UserCollection, id, bson.M{"$pullAll": bson.M{"abilities": []string{"read"}}})
	})
}

func TestGetUsersWithAbility(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll(UserCollection).Docs = []any{User{Username: "a"}, User{Username: "b"}}

	users, err := GetUsersWithAbility(context.Background(), "read", db)
	if err != nil || len(users) != 2 || users[1].Username != "b" {
		t.Errorf("got %+v, %v", users, err)
	}
	checkCall(t, db, UserCollection, "Find", bson.M{"abilities": "read"})
}
