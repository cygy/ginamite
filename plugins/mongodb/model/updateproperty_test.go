package model

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestUpdatePropertyProcess_Document(t *testing.T) {
	testDocument(t, UpdatePropertyProcessCollection, func() document.Document { return &UpdatePropertyProcess{} })
}

func TestNewUpdatePropertyProcess(t *testing.T) {
	id, hex := savedID()

	process, err := NewUpdatePropertyProcess(hex, "type", "value")
	if err != nil {
		t.Fatal(err)
	}
	if process.UserID != id || process.Type != "type" || process.Value != "value" || process.Tries != 0 || process.TTLKey != hex+"type" || len(process.Code) != 64 {
		t.Errorf("got %+v", process)
	}

	other, _ := NewUpdatePropertyProcess(hex, "type", "value")
	if other.Code == process.Code {
		t.Error("each process needs its own code")
	}

	if _, err := NewUpdatePropertyProcess("nope", "type", "value"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestNewProcesses(t *testing.T) {
	_, hex := savedID()

	tests := []struct {
		name      string
		create    func() (*UpdatePropertyProcess, error)
		wantType  string
		wantValue string
	}{
		{"password", func() (*UpdatePropertyProcess, error) { return NewUpdatePasswordProcess(hex, "hash") }, authentication.ProcessUpdatePassword, "hash"},
		{"email address", func() (*UpdatePropertyProcess, error) { return NewUpdateEmailAddressProcess(hex, "a@example.com") }, authentication.ProcessUpdateEmailAddress, "a@example.com"},
		{"valid email address", func() (*UpdatePropertyProcess, error) { return NewUpdateValidEmailAddressProcess(hex) }, authentication.ProcessVerifyEmailAddress, "true"},
		{"delete account", func() (*UpdatePropertyProcess, error) { return NewDeleteAccountProcess(hex) }, authentication.ProcessDeleteAccount, "true"},
		{"disable account", func() (*UpdatePropertyProcess, error) { return NewDisableAccountProcess(hex) }, authentication.ProcessDisableAccount, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			process, err := tt.create()
			if err != nil || process.Type != tt.wantType || process.Value != tt.wantValue {
				t.Errorf("got %+v, %v", process, err)
			}
		})
	}
}

func TestUpdatePropertyProcess_Get(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	tests := []struct {
		name     string
		call     func(p *UpdatePropertyProcess, db *mock.Database) error
		wantType string
	}{
		{"by type", func(p *UpdatePropertyProcess, db *mock.Database) error {
			return p.GetByUserIDAndType(ctx, hex, "custom", db)
		}, "custom"},
		{"password", func(p *UpdatePropertyProcess, db *mock.Database) error {
			return p.GetUpdatePasswordProcessByUserID(ctx, hex, db)
		}, authentication.ProcessUpdatePassword},
		{"email address", func(p *UpdatePropertyProcess, db *mock.Database) error {
			return p.GetUpdateEmailAddressProcessByUserID(ctx, hex, db)
		}, authentication.ProcessUpdateEmailAddress},
		{"valid email address", func(p *UpdatePropertyProcess, db *mock.Database) error {
			return p.GetUpdateValidEmailAddressProcessByUserID(ctx, hex, db)
		}, authentication.ProcessVerifyEmailAddress},
		{"delete account", func(p *UpdatePropertyProcess, db *mock.Database) error {
			return p.GetDeleteAccountProcessByUserID(ctx, hex, db)
		}, authentication.ProcessDeleteAccount},
		{"disable account", func(p *UpdatePropertyProcess, db *mock.Database) error {
			return p.GetDisableAccountProcessByUserID(ctx, hex, db)
		}, authentication.ProcessDisableAccount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			db.Coll(UpdatePropertyProcessCollection).One = mock.Result{Value: bson.M{"_id": mustObjectID(), "code": "abc"}}

			process := &UpdatePropertyProcess{}
			if err := tt.call(process, db); err != nil || process.Code != "abc" {
				t.Fatalf("got %+v, %v", process, err)
			}
			checkCall(t, db, UpdatePropertyProcessCollection, "FindOne", bson.D{{Key: "userId", Value: id}, {Key: "type", Value: tt.wantType}})

			db.Coll(UpdatePropertyProcessCollection).Err = errBoom
			if err := tt.call(process, db); !errors.Is(err, errBoom) {
				t.Errorf("err = %v", err)
			}
		})
	}

	if err := (&UpdatePropertyProcess{}).GetByUserIDAndType(ctx, "nope", "t", mock.NewDatabase("app")); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestUpdatePropertyProcess_Increment(t *testing.T) {
	id, hex := savedID()
	process := &UpdatePropertyProcess{}
	_ = process.SetID(hex)

	db := mock.NewDatabase("app")
	if err := process.Increment(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, UpdatePropertyProcessCollection, id, bson.M{"$inc": bson.M{"tries": 1}})
}

func TestDeleteUpdatePropertyProcesses(t *testing.T) {
	ctx := context.Background()
	id, hex := savedID()

	tests := []struct {
		name     string
		call     func(db *mock.Database) error
		wantType string
	}{
		{"by type", func(db *mock.Database) error {
			return DeleteUpdatePropertyProcessByUserIDAndType(ctx, hex, "custom", db)
		}, "custom"},
		{"password", func(db *mock.Database) error { return DeleteUpdatePasswordProcessByUserID(ctx, hex, db) }, authentication.ProcessUpdatePassword},
		{"email address", func(db *mock.Database) error { return DeleteUpdateEmailAddressProcessByUserID(ctx, hex, db) }, authentication.ProcessUpdateEmailAddress},
		{"valid email address", func(db *mock.Database) error { return DeleteUpdateValidEmailAddressProcessByUserID(ctx, hex, db) }, authentication.ProcessVerifyEmailAddress},
		{"delete account", func(db *mock.Database) error { return DeleteDeleteAccountProcessByUserID(ctx, hex, db) }, authentication.ProcessDeleteAccount},
		{"disable account", func(db *mock.Database) error { return DeleteDisableAccountProcessByUserID(ctx, hex, db) }, authentication.ProcessDisableAccount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := mock.NewDatabase("app")
			if err := tt.call(db); err != nil {
				t.Fatal(err)
			}
			checkCall(t, db, UpdatePropertyProcessCollection, "DeleteMany", bson.D{{Key: "userId", Value: id}, {Key: "type", Value: tt.wantType}})
		})
	}

	if err := DeleteUpdatePropertyProcessByUserIDAndType(ctx, "nope", "t", mock.NewDatabase("app")); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}
