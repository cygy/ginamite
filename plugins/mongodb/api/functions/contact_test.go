package functions

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

var ctx = context.Background()

func TestContacts_GetContacts(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll(model.ContactCollection).Docs = []any{model.Contact{Subject: "one"}}
	contacts := NewContacts(func() store.Database { return db })

	got, err := contacts.GetContacts(ctx, 0, 10)
	if contacts, ok := got.([]model.AdminSummarizedContact); err != nil || !ok || len(contacts) != 1 || contacts[0].Subject != "one" {
		t.Errorf("got %+v, %v", got, err)
	}

	db.Coll(model.ContactCollection).Err = errBoom
	if _, err := contacts.GetContacts(ctx, 0, 10); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContacts_GetContact(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.ContactCollection).One = mock.Result{Value: bson.M{"_id": id, "subject": "one", "text": "long"}}
	contacts := NewContacts(func() store.Database { return db })

	got, err := contacts.GetContact(ctx, id.Hex())
	if contact, ok := got.(*model.AdminContact); err != nil || !ok || contact.Subject != "one" || contact.Text != "long" {
		t.Errorf("got %+v, %v", got, err)
	}

	if _, err := contacts.GetContact(ctx, "nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}

func TestContacts_UpdateContact(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.ContactCollection).One = mock.Result{Value: bson.M{"_id": id}}
	contacts := NewContacts(func() store.Database { return db })

	if err := contacts.UpdateContact(ctx, id.Hex(), true); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.ContactCollection, id, bson.M{"$set": bson.M{"done": true}})

	// A contact that does not exist is not updated.
	db = mock.NewDatabase("app")
	if err := contacts.UpdateContact(ctx, id.Hex(), true); err == nil {
		t.Error("an unknown contact must be reported")
	}
	if got := calls(db, model.ContactCollection); len(got) != 1 {
		t.Errorf("calls = %v, want only the lookup", got)
	}
}

func TestContacts_CreateContact(t *testing.T) {
	db := mock.NewDatabase("app")
	contacts := NewContacts(func() store.Database { return db })

	if err := contacts.CreateContact(ctx, "j@example.com", "subject", "text"); err != nil {
		t.Fatal(err)
	}

	saved := db.Coll(model.ContactCollection).LastCall().Args[0].(*model.Contact)
	if saved.EmailAddress != "j@example.com" || saved.Subject != "subject" || saved.Text != "text" || saved.Done {
		t.Errorf("contact = %+v", saved)
	}

	db.Coll(model.ContactCollection).Err = errBoom
	if err := contacts.CreateContact(ctx, "a", "b", "c"); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}
