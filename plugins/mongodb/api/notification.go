// Package api provides the helpers of the MongoDB plugin shared by the API functions.
package api

import (
	"context"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

// CreateStoredNotification saves a notification for a user, and updates the counts of notifications of the user.
// The oldest notifications are deleted when the user has more than limit notifications.
// The image is kept only if it has a URL. Failing to delete the extra notifications or to update the counts is only
// logged.
func CreateStoredNotification(ctx context.Context, userID, notificationType, data string, image model.PublicImage, locales map[string]model.NotificationLocale, limit int64, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Create and save the user notifications.
	notification := model.NewNotification()
	notification.UserID = objectID
	notification.Type = notificationType
	notification.Data = data
	notification.Locales = locales
	if len(image.Full.Size1x) > 0 {
		notification.Image = image
	}

	if _, err := notification.Save(ctx, db); err != nil {
		return err
	}

	// Get the count of user notifications in order to delete the extra notifications if the count is superior to the limit.
	countOfNotifications, err := model.GetCountOfUserNotifications(ctx, userID, db)
	if err != nil {
		return err
	}
	if countOfNotifications > limit {
		notificationIDs, err := model.GetOldestUserNotifications(ctx, userID, limit, db)
		if err == nil {
			if _, err := model.DeleteUserNotifications(ctx, notificationIDs, db); err != nil {
				log.WithField("error", err).Error("unable to delete user notifications")
			}
		}

		countOfNotifications, _ = model.GetCountOfUserNotifications(ctx, userID, db)
	}

	// Update the count of user notifications, total and unread.
	countOfUnreadNotifications, err := model.GetCountOfUserUnreadNotifications(ctx, userID, db)
	if err != nil {
		return err
	}

	user := model.User{}
	user.ID = notification.UserID
	if err := user.UpdateCountOfNotifications(ctx, countOfNotifications, countOfUnreadNotifications, db); err != nil {
		log.WithField("error", err).Error("unable to update count of notifications")
	}

	return nil
}
