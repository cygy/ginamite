package model

import (
	"testing"
	"time"
)

func TestNewOwnAuthToken(t *testing.T) {
	now := time.Now()
	token := &AuthToken{
		ID: mustObjectID(), Source: "web", Method: "password", CreatedAt: now, ExpiresAt: now.Add(time.Hour),
		Device:   Device{Name: "phone", NotificationSettings: DeviceNotification{Enabled: true, Token: "secret"}},
		Location: &IPLocation{IP: "1.2.3.4", City: "Paris", Region: "IDF"},
	}

	got := NewOwnAuthToken(token)
	if got.ID != token.ID || got.Source != "web" || got.Method != "password" || !got.Device.NotificationSettings.Enabled || got.Location == nil || got.Location.City != "Paris" {
		t.Errorf("got %+v", got)
	}
	if !got.HasLocation() || got.URLParams()["id"] != token.ID.Hex() {
		t.Errorf("HasLocation() = %v, URLParams() = %v", got.HasLocation(), got.URLParams())
	}

	token.Location = nil
	if got := NewOwnAuthToken(token); got.Location != nil || got.HasLocation() {
		t.Error("a token without location")
	}

	token.Location = &IPLocation{}
	if NewOwnAuthToken(token).HasLocation() {
		t.Error("an empty location is not a location")
	}
}
