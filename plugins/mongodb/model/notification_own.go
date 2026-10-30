package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
)

// OwnSummarizedNotification is the short view of a Notification exposed to its owner.
type OwnSummarizedNotification struct {
	ID       bson.ObjectID `json:"id,omitempty"`
	Type     string        `json:"type"`
	Image    OwnImageSize  `json:"image,omitempty"`
	Title    string        `json:"title"`
	Preview  string        `json:"preview"`
	Read     bool          `json:"read"`
	Notified bool          `json:"notified"`
}

// NewOwnSummarizedNotification returns a new OwnSummarizedNotification.
func NewOwnSummarizedNotification(notification Notification, locale, defaultLocale string) OwnSummarizedNotification {
	var title, preview string
	if value, ok := notification.Locales[locale]; ok {
		title = value.Title
		preview = value.Preview
	} else if value, ok := notification.Locales[defaultLocale]; ok {
		title = value.Title
		preview = value.Preview
	}

	return OwnSummarizedNotification{
		ID:   notification.ID,
		Type: notification.Type,
		Image: OwnImageSize{
			Size1x: notification.Image.Small.Size1x,
			Size2x: notification.Image.Small.Size2x,
		},
		Title:    title,
		Preview:  preview,
		Read:     notification.Read,
		Notified: notification.Notified,
	}
}

// URLParams returns the parameters to build the URL of the document: its ID.
func (notification OwnSummarizedNotification) URLParams() map[string]string {
	return map[string]string{
		"id": notification.ID.Hex(),
	}
}

// OwnNotification is the view of a Notification exposed to its owner.
type OwnNotification struct {
	ID        bson.ObjectID `json:"id,omitempty"`
	Type      string        `json:"type"`
	Data      string        `json:"data"`
	Image     OwnImageSize  `json:"image,omitempty"`
	Title     string        `json:"title"`
	Text      string        `json:"text"`
	Read      bool          `json:"read"`
	Notified  bool          `json:"notified"`
	CreatedAt time.Time     `json:"created_at"`
}

// NewOwnNotification returns a new OwnNotification.
func NewOwnNotification(notification Notification, locale, defaultLocale string) OwnNotification {
	var title, text string
	if value, ok := notification.Locales[locale]; ok {
		title = value.Title
		text = value.Text
	} else if value, ok := notification.Locales[defaultLocale]; ok {
		title = value.Title
		text = value.Text
	}

	return OwnNotification{
		ID:   notification.ID,
		Type: notification.Type,
		Data: notification.Data,
		Image: OwnImageSize{
			Size1x: notification.Image.Full.Size1x,
			Size2x: notification.Image.Full.Size2x,
		},
		Title:     title,
		Text:      text,
		Read:      notification.Read,
		Notified:  notification.Notified,
		CreatedAt: notification.CreatedAt,
	}
}

// URLParams returns the parameters to build the URL of the document: its ID.
func (notification OwnNotification) URLParams() map[string]string {
	return map[string]string{
		"id": notification.ID.Hex(),
	}
}
