package model

import (
	"testing"
	"time"
)

func TestNewAdminContact(t *testing.T) {
	now := time.Now()
	contact := &Contact{ID: mustObjectID(), EmailAddress: "a@example.com", Subject: "s", Text: "t", Done: true, CreatedAt: now}

	got := NewAdminContact(contact)
	if got.ID != contact.ID || got.EmailAddress != "a@example.com" || got.Subject != "s" || got.Text != "t" || !got.Done || !got.CreatedAt.Equal(now) {
		t.Errorf("got %+v", got)
	}
	if got.URLParams()["id"] != contact.ID.Hex() {
		t.Errorf("URLParams() = %v", got.URLParams())
	}
}

func TestNewAdminSummarizedContact(t *testing.T) {
	now := time.Now()
	contact := &Contact{ID: mustObjectID(), EmailAddress: "a@example.com", Subject: "s", Text: "t", Done: true, CreatedAt: now}

	got := NewAdminSummarizedContact(contact)
	if got.ID != contact.ID || got.EmailAddress != "a@example.com" || got.Subject != "s" || !got.Done || !got.CreatedAt.Equal(now) {
		t.Errorf("got %+v", got)
	}
	if got.URLParams()["id"] != contact.ID.Hex() {
		t.Errorf("URLParams() = %v", got.URLParams())
	}
}
