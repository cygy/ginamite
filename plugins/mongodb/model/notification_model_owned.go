package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// OwnSummarizedNotification : own version of the Notification struct
type OwnSummarizedNotification struct {
	ID       bson.ObjectId `json:"id,omitempty"`
	Type     string        `json:"type"`
	Image    OwnImageSize  `json:"image,omitempty"`
	Title    string        `json:"title"`
	Preview  string        `json:"preview"`
	Read     bool          `json:"read"`
	Notified bool          `json:"notified"`
}

// NewOwnSummarizedNotification : returns a new 'OwnSummarizedNotification' struct.
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

// URLParams : returns a struct containing the parameters to use into a url.
func (notification OwnSummarizedNotification) URLParams() map[string]string {
	return map[string]string{
		"id": notification.ID.Hex(),
	}
}

// OwnNotification : own version of the Notification struct
type OwnNotification struct {
	ID        bson.ObjectId `json:"id,omitempty"`
	Type      string        `json:"type"`
	Data      string        `json:"data"`
	Image     OwnImageSize  `json:"image,omitempty"`
	Title     string        `json:"title"`
	Text      string        `json:"text"`
	Read      bool          `json:"read"`
	Notified  bool          `json:"notified"`
	CreatedAt time.Time     `json:"created_at"`
}

// NewOwnNotification : returns a new 'OwnNotification' struct.
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

// URLParams : returns a struct containing the parameters to use into a url.
func (notification OwnNotification) URLParams() map[string]string {
	return map[string]string{
		"id": notification.ID.Hex(),
	}
}
