package model

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Notification : notification properties.
type Notification struct {
	ID        bson.ObjectId                 `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    bson.ObjectId                 `json:"user_id" bson:"userId"`
	Type      string                        `json:"type" bson:"type"`
	Data      string                        `json:"data" bson:"data"`
	Image     PublicImage                   `json:"image,omitempty" bson:"image,omitempty"`
	Locales   map[string]NotificationLocale `json:"locales" bson:"locales"`
	Read      bool                          `json:"read" bson:"read"`
	Notified  bool                          `json:"notified" bson:"notified"`
	CreatedAt time.Time                     `json:"created_at" bson:"createdAt"`
}

// NotificationLocale : struct of a locale of a notification.
type NotificationLocale struct {
	Title   string `json:"title" bson:"title"`
	Text    string `json:"text" bson:"text"`       // long description
	Preview string `json:"preview" bson:"preview"` // short description
}

// NewNotification : returns a new 'Notification' struct.
func NewNotification() *Notification {
	return &Notification{
		Read:      false,
		Notified:  false,
		CreatedAt: time.Now(),
	}
}

// IsSaved : returns true if the 'Notification' document if saved in the collection.
func (notification *Notification) IsSaved() bool {
	return notification.ID.Valid()
}

// Save : inserts the 'Notification' document in the collection, returns an error if needed.
func (notification *Notification) Save(db *mgo.Database) (string, error) {
	return document.SaveDocument(notification, NotificationCollection, db)
}

// Update : updates the 'Notification' document, returns an error if needed.
func (notification *Notification) Update(update interface{}, db *mgo.Database) error {
	return document.UpdateDocument(notification.ID, NotificationCollection, update, db)
}

// Delete : deletes the 'Notification' document from the collection, returns an error if needed.
func (notification *Notification) Delete(db *mgo.Database) error {
	return document.DeleteDocument(notification.ID, NotificationCollection, db)
}

// GetByID : initializes the 'Notification' struct from the 'Notification' document with the ID.
func (notification *Notification) GetByID(id string, db *mgo.Database) error {
	return document.GetDocumentByID(notification, NotificationCollection, id, db)
}

// GetID : returns the 'Notification' ID.
func (notification *Notification) GetID() string {
	return notification.ID.Hex()
}

// SetID : initializes the 'Notification' ID.
func (notification *Notification) SetID(id string) {
	notification.ID = bson.ObjectIdHex(id)
}
