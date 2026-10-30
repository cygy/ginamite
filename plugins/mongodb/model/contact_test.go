package model

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestContact_Document(t *testing.T) {
	testDocument(t, ContactCollection, func() document.Document { return NewContact() })
}

func TestNewContact(t *testing.T) {
	contact := NewContact()
	if contact.Done || contact.CreatedAt.IsZero() {
		t.Errorf("contact = %+v", contact)
	}
}

func TestContact_UpdateDone(t *testing.T) {
	id, hex := savedID()
	contact := NewContact()
	_ = contact.SetID(hex)

	db := mock.NewDatabase("app")
	if err := contact.UpdateDone(context.Background(), true, db); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, ContactCollection, id, bson.M{"$set": bson.M{"done": true}})
}

func TestGetContactsForAdmin(t *testing.T) {
	ctx := context.Background()
	db := mock.NewDatabase("app")
	db.Coll(ContactCollection).Docs = []any{
		Contact{EmailAddress: "a@example.com", Subject: "one", Text: "long text"},
		Contact{EmailAddress: "b@example.com", Subject: "two", Done: true},
	}

	contacts, err := GetContactsForAdmin(ctx, 10, 20, db)
	if err != nil || len(contacts) != 2 {
		t.Fatalf("got %+v, %v", contacts, err)
	}
	if contacts[0].EmailAddress != "a@example.com" || contacts[1].Subject != "two" || !contacts[1].Done {
		t.Errorf("contacts = %+v", contacts)
	}

	opts := mock.FindOptions(lastCall(db, ContactCollection).Opts)
	if *opts.Skip != 10 || *opts.Limit != 20 || opts.Sort == nil {
		t.Errorf("options = %+v", opts)
	}

	db.Coll(ContactCollection).Err = errBoom
	if contacts, err := GetContactsForAdmin(ctx, 0, 1, db); !errors.Is(err, errBoom) || len(contacts) != 0 {
		t.Errorf("got %+v, %v", contacts, err)
	}
}
