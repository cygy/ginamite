package model

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestForgotPasswordProcess_Document(t *testing.T) {
	testDocument(t, ForgotPasswordProcessCollection, func() document.Document { return &ForgotPasswordProcess{} })
}

func TestNewForgotPasswordProcess(t *testing.T) {
	id, hex := savedID()

	process, err := NewForgotPasswordProcess(hex)
	if err != nil || process.UserID != id || process.Tries != 0 {
		t.Errorf("got %+v, %v", process, err)
	}

	if _, err := NewForgotPasswordProcess("nope"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestForgotPasswordProcess_GetByUserID(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	db := mock.NewDatabase("app")
	db.Coll(ForgotPasswordProcessCollection).One = mock.Result{Value: bson.M{"_id": mustObjectID(), "userId": id, "code": "abc"}}

	process := &ForgotPasswordProcess{}
	if err := process.GetByUserID(ctx, hex, db); err != nil || process.Code != "abc" || process.UserID != id {
		t.Errorf("got %+v, %v", process, err)
	}
	checkCall(t, db, ForgotPasswordProcessCollection, "FindOne", bson.M{"userId": id})

	if err := process.GetByUserID(ctx, "nope", db); err == nil {
		t.Error("an invalid user ID must be rejected")
	}

	db.Coll(ForgotPasswordProcessCollection).Err = errBoom
	if err := process.GetByUserID(ctx, hex, db); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestForgotPasswordProcess_Increment(t *testing.T) {
	id, hex := savedID()
	process := &ForgotPasswordProcess{}
	_ = process.SetID(hex)

	db := mock.NewDatabase("app")
	if err := process.Increment(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, ForgotPasswordProcessCollection, id, bson.M{"$inc": bson.M{"tries": 1}})
}

func TestDeleteForgotPasswordProcessByUserID(t *testing.T) {
	id, hex := savedID()

	db := mock.NewDatabase("app")
	if err := DeleteForgotPasswordProcessByUserID(context.Background(), hex, db); err != nil {
		t.Fatal(err)
	}
	checkCall(t, db, ForgotPasswordProcessCollection, "DeleteMany", bson.M{"userId": id})

	if err := DeleteForgotPasswordProcessByUserID(context.Background(), "nope", db); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}
