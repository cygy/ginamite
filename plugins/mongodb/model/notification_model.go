package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Notification : notification properties.
type Notification struct {
	ID        bson.ObjectID                 `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    bson.ObjectID                 `json:"user_id" bson:"userId"`
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
	return bson.IsObjectIDProvided(notification.ID)
}

// Save : inserts the 'Notification' document in the collection, returns an error if needed.
func (notification *Notification) Save(ctx context.Context, db *mongo.Database) (string, error) {
	return document.SaveDocument(ctx, notification, NotificationCollection, db)
}

// Update : updates the 'Notification' document, returns an error if needed.
func (notification *Notification) Update(ctx context.Context, update any, db *mongo.Database) error {
	return document.UpdateDocument(ctx, notification.ID, NotificationCollection, update, db)
}

// Delete : deletes the 'Notification' document from the collection, returns an error if needed.
func (notification *Notification) Delete(ctx context.Context, db *mongo.Database) error {
	return document.DeleteDocument(ctx, notification.ID, NotificationCollection, db)
}

// GetByID : initializes the 'Notification' struct from the 'Notification' document with the ID.
func (notification *Notification) GetByID(ctx context.Context, id string, db *mongo.Database) error {
	return document.GetDocumentByID(ctx, notification, NotificationCollection, id, db)
}

// GetID : returns the 'Notification' ID.
func (notification *Notification) GetID() string {
	return notification.ID.Hex()
}

// SetID : initializes the 'Notification' ID.
func (notification *Notification) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	notification.ID = objectID
	return nil
}
