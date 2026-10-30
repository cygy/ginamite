package model

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestDeletedUser_Document(t *testing.T) {
	testDocument(t, DeletedUserCollection, func() document.Document { return &DeletedUser{} })
}

func TestNewDeletedUser(t *testing.T) {
	id, hex := savedID()

	user, err := NewDeletedUser(hex)
	if err != nil || user.UserID != id || user.CreatedAt.IsZero() {
		t.Errorf("got %+v, %v", user, err)
	}

	if _, err := NewDeletedUser("nope"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestDeleteDeletedUserByUserID(t *testing.T) {
	id, hex := savedID()

	db := mock.NewDatabase("app")
	if err := DeleteDeletedUserByUserID(context.Background(), hex, db); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, DeletedUserCollection, "DeleteMany", bson.M{"userId": id})

	if err := DeleteDeletedUserByUserID(context.Background(), "nope", db); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}
