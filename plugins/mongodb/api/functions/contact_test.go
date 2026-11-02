package functions

import (
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestGetContactsForAdmin(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll(model.ContactCollection).Docs = []any{model.Contact{Subject: "one"}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	got, err := GetContactsForAdmin(c, 0, 10)
	if contacts, ok := got.([]model.AdminSummarizedContact); err != nil || !ok || len(contacts) != 1 || contacts[0].Subject != "one" {
		t.Errorf("got %+v, %v", got, err)
	}

	db.Coll(model.ContactCollection).Err = errBoom
	if _, err := GetContactsForAdmin(c, 0, 10); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestGetContactForAdmin(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.ContactCollection).One = mock.Result{Value: bson.M{"_id": id, "subject": "one", "text": "long"}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	got, err := GetContactForAdmin(c, id.Hex())
	if contact, ok := got.(*model.AdminContact); err != nil || !ok || contact.Subject != "one" || contact.Text != "long" {
		t.Errorf("got %+v, %v", got, err)
	}

	if _, err := GetContactForAdmin(c, "nope"); err == nil {
		t.Error("an invalid ID must be rejected")
	}
}

func TestUpdateContact(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.ContactCollection).One = mock.Result{Value: bson.M{"_id": id}}
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := UpdateContact(c, id.Hex(), true); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.ContactCollection, id, bson.M{"$set": bson.M{"done": true}})

	// A contact that does not exist is not updated.
	db = mock.NewDatabase("app")
	c, _ = plugintest.NewContext(db, "GET", "/", nil)
	if err := UpdateContact(c, id.Hex(), true); err == nil {
		t.Error("an unknown contact must be reported")
	}
	if got := calls(db, model.ContactCollection); len(got) != 1 {
		t.Errorf("calls = %v, want only the lookup", got)
	}
}

func TestCreateContact(t *testing.T) {
	db := mock.NewDatabase("app")
	c, _ := plugintest.NewContext(db, "GET", "/", nil)

	if err := CreateContact(c, "j@example.com", "subject", "text"); err != nil {
		t.Fatal(err)
	}

	saved := db.Coll(model.ContactCollection).LastCall().Args[0].(*model.Contact)
	if saved.EmailAddress != "j@example.com" || saved.Subject != "subject" || saved.Text != "text" || saved.Done {
		t.Errorf("contact = %+v", saved)
	}

	db.Coll(model.ContactCollection).Err = errBoom
	if err := CreateContact(c, "a", "b", "c"); !errors.Is(err, errBoom) {
		t.Errorf("err = %v", err)
	}
}
