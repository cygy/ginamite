package model

import "testing"

func TestNewAdminDisabledUser(t *testing.T) {
	user := &User{ID: mustObjectID(), Username: "john", Slug: "john", Verified: true, PrivateInfos: UserPrivateInfos{Email: "john@example.com"}}

	got := NewAdminDisabledUser(user)
	if got.Username != "john" || got.Slug != "john" || !got.Verified || got.PrivateInfos.Email != "john@example.com" {
		t.Errorf("got %+v", got)
	}

	// The ID of the view is the ID of the disabled document, set by the query.
	got.ID = user.ID
	if got.URLParams()["id"] != user.ID.Hex() {
		t.Errorf("URLParams() = %v", got.URLParams())
	}
}
