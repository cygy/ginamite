package model

import (
	"testing"
	"time"
)

func testNotification() Notification {
	return Notification{
		ID: mustObjectID(), Type: "news", Data: "data", Read: true, Notified: true, CreatedAt: time.Now(),
		Image: PublicImage{
			Full:  PublicImageSize{Size1x: "full1", Size2x: "full2"},
			Small: PublicImageSize{Size1x: "small1", Size2x: "small2"},
		},
		Locales: map[string]NotificationLocale{
			"en": {Title: "Title", Text: "Text", Preview: "Preview"},
			"fr": {Title: "Titre", Text: "Texte", Preview: "Aperçu"},
		},
	}
}

func TestNewOwnSummarizedNotification(t *testing.T) {
	notification := testNotification()

	tests := []struct {
		name, locale, wantTitle, wantPreview string
	}{
		{"locale", "fr", "Titre", "Aperçu"},
		{"default locale", "de", "Title", "Preview"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOwnSummarizedNotification(notification, tt.locale, "en")
			if got.Title != tt.wantTitle || got.Preview != tt.wantPreview {
				t.Errorf("got %q, %q", got.Title, got.Preview)
			}
			if got.ID != notification.ID || got.Type != "news" || !got.Read || !got.Notified || got.Image.Size1x != "small1" || got.Image.Size2x != "small2" {
				t.Errorf("got %+v", got)
			}
			if got.URLParams()["id"] != notification.ID.Hex() {
				t.Errorf("URLParams() = %v", got.URLParams())
			}
		})
	}

	if got := NewOwnSummarizedNotification(notification, "de", "es"); got.Title != "" || got.Preview != "" {
		t.Errorf("without matching locale, got %q, %q", got.Title, got.Preview)
	}
}

func TestNewOwnNotification(t *testing.T) {
	notification := testNotification()

	tests := []struct {
		name, locale, wantTitle, wantText string
	}{
		{"locale", "fr", "Titre", "Texte"},
		{"default locale", "de", "Title", "Text"},
		{"unknown", "de", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultLocale := "en"
			if tt.name == "unknown" {
				defaultLocale = "es"
			}

			got := NewOwnNotification(notification, tt.locale, defaultLocale)
			if got.Title != tt.wantTitle || got.Text != tt.wantText {
				t.Errorf("got %q, %q", got.Title, got.Text)
			}
			if got.ID != notification.ID || got.Data != "data" || !got.Read || !got.Notified || got.Image.Size1x != "full1" || got.Image.Size2x != "full2" || !got.CreatedAt.Equal(notification.CreatedAt) {
				t.Errorf("got %+v", got)
			}
			if got.URLParams()["id"] != notification.ID.Hex() {
				t.Errorf("URLParams() = %v", got.URLParams())
			}
		})
	}
}
