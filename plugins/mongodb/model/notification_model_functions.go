package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UpdateRead : updated the read property of a notification received by a user.
func (notification *Notification) UpdateRead(ctx context.Context, db store.Database) error {
	return notification.Update(ctx, bson.M{"$set": bson.M{"read": true}}, db)
}

// UpdateNotified : updated the notified property of a notification received by a user.
func (notification *Notification) UpdateNotified(ctx context.Context, db store.Database) error {
	return notification.Update(ctx, bson.M{"$set": bson.M{"notified": true}}, db)
}
