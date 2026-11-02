package model

import (
	"testing"
	"time"
)

func TestNewAdminAuthToken(t *testing.T) {
	now := time.Now()
	token := &AuthToken{
		ID: mustObjectID(), Source: "web", Method: "password", CreatedAt: now, ExpiresAt: now.Add(time.Hour),
		Device:   Device{Name: "phone", Details: "d", Type: "mobile"},
		Location: &IPLocation{IP: "1.2.3.4", Region: "IDF", Country: IPLocationCountry{Name: "France"}},
	}

	got := NewAdminAuthToken(token)
	if got.ID != token.ID || got.Source != "web" || got.Method != "password" || got.Device.Name != "phone" || got.Location == nil || got.Location.IP != "1.2.3.4" {
		t.Errorf("got %+v", got)
	}
	if !got.CreatedAt.Equal(now) || !got.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Errorf("dates = %v, %v", got.CreatedAt, got.ExpiresAt)
	}
	if !got.HasLocation() {
		t.Error("the token has a location")
	}
	if got.URLParams()["id"] != token.ID.Hex() {
		t.Errorf("URLParams() = %v", got.URLParams())
	}

	token.Location = nil
	if got := NewAdminAuthToken(token); got.Location != nil || got.HasLocation() {
		t.Errorf("a token without location, got %+v", got.Location)
	}

	token.Location = &IPLocation{}
	if NewAdminAuthToken(token).HasLocation() {
		t.Error("an empty location is not a location")
	}
}
