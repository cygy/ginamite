package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// NotificationCollection is the name of the collection of the notifications documents.
const NotificationCollection = "notifications"

// Notification is a notification received by a user.
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

// NotificationLocale is the text of a Notification in a locale.
type NotificationLocale struct {
	Title   string `json:"title" bson:"title"`
	Text    string `json:"text" bson:"text"`       // long description
	Preview string `json:"preview" bson:"preview"` // short description
}

// NewNotification returns a new Notification.
func NewNotification() *Notification {
	return &Notification{
		Read:      false,
		Notified:  false,
		CreatedAt: time.Now(),
	}
}

// IsSaved reports whether the Notification document is saved in the collection.
func (notification *Notification) IsSaved() bool {
	return bson.IsObjectIDProvided(notification.ID)
}

// Save inserts the Notification document in the collection, returning an error if needed.
func (notification *Notification) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, notification, NotificationCollection, db)
}

// Update updates the Notification document, returning an error if needed.
func (notification *Notification) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, notification.ID, NotificationCollection, update, db)
}

// Delete deletes the Notification document from the collection, returning an error if needed.
func (notification *Notification) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, notification.ID, NotificationCollection, db)
}

// GetByID loads the Notification document with the ID.
func (notification *Notification) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, notification, NotificationCollection, id, db)
}

// GetID returns the Notification ID.
func (notification *Notification) GetID() string {
	return notification.ID.Hex()
}

// SetID sets the ID of the Notification from its hexadecimal form. The ID is kept if the form is not valid.
func (notification *Notification) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	notification.ID = objectID
	return nil
}

// UpdateRead updates the read property of a notification received by a user.
func (notification *Notification) UpdateRead(ctx context.Context, db store.Database) error {
	return notification.Update(ctx, bson.M{"$set": bson.M{"read": true}}, db)
}

// UpdateNotified updates the notified property of a notification received by a user.
func (notification *Notification) UpdateNotified(ctx context.Context, db store.Database) error {
	return notification.Update(ctx, bson.M{"$set": bson.M{"notified": true}}, db)
}

// GetUserNotifications returns the latest notifications received by a user.
func GetUserNotifications(ctx context.Context, userID string, db store.Database) ([]Notification, error) {
	notifications := []Notification{}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return notifications, err
	}

	err = document.GetDocumentsBySelectorAndSort(ctx, &notifications, NotificationCollection, bson.M{"userId": objectID}, []string{"userId", "-createdAt"}, db)
	return notifications, err
}

// GetCountOfUserNotifications returns the count of notifications received by a user.
func GetCountOfUserNotifications(ctx context.Context, userID string, db store.Database) (int64, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	return document.CountBySelector(ctx, NotificationCollection, bson.M{"userId": objectID}, db)
}

// GetCountOfUserUnreadNotifications returns the count of unread notifications received by a user.
func GetCountOfUserUnreadNotifications(ctx context.Context, userID string, db store.Database) (int64, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	return document.CountBySelector(ctx, NotificationCollection, bson.D{
		{Key: "userId", Value: objectID},
		{Key: "read", Value: false},
	}, db)
}

// GetOldestUserNotifications returns the IDs of the oldest notifications received by a user.
func GetOldestUserNotifications(ctx context.Context, userID string, offset int64, db store.Database) ([]bson.ObjectID, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return []bson.ObjectID{}, err
	}

	notifications := []Notification{}
	err = document.GetDocumentsBySelectorAndSortWithOffsetAndLimit(ctx, &notifications, NotificationCollection, bson.M{"userId": objectID}, []string{"userId", "-createdAt"}, offset, -1, db)

	notificationIDs := make([]bson.ObjectID, len(notifications))
	for i, notification := range notifications {
		notificationIDs[i] = notification.ID
	}

	return notificationIDs, err
}

// DeleteUserNotifications deletes the notifications with these IDs and returns the number of notifications deleted.
func DeleteUserNotifications(ctx context.Context, notificationsIDs []bson.ObjectID, db store.Database) (int64, error) {
	// MongoDB rejects an empty $or.
	if len(notificationsIDs) == 0 {
		return 0, nil
	}

	IDs := make([]bson.M, len(notificationsIDs))
	for i, notificationsID := range notificationsIDs {
		IDs[i] = bson.M{"_id": notificationsID}
	}

	return document.DeleteDocumentsBySelector(ctx, NotificationCollection, bson.M{"$or": IDs}, db)
}
