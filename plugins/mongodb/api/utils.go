package api

import (
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// CreateStoredNotification : creates a notification for a user.
func CreateStoredNotification(userID, notificationType, data string, image model.PublicImage, locales map[string]model.NotificationLocale, db *mgo.Database) error {
	// Create and save the user notifications.
	notification := model.NewNotification()
	notification.UserID = bson.ObjectIdHex(userID)
	notification.Type = notificationType
	notification.Data = data
	notification.Locales = locales
	if len(image.Full.Size1x) > 0 {
		notification.Image = image
	}

	if _, err := notification.Save(db); err != nil {
		return err
	}

	// Get the count of user notifications in order to delete the extra notifications if the count is superior to the limit.
	countOfNotifications, err := model.GetCountOfUserNotifications(userID, db)
	if err != nil {
		return err
	}
	if countOfNotifications > config.Main.UserNotifications.Limit {
		notificationIDs, err := model.GetOldestUserNotifications(userID, config.Main.UserNotifications.Limit, db)
		if err == nil {
			model.DeleteUserNotifications(notificationIDs, db)
		}

		countOfNotifications, _ = model.GetCountOfUserNotifications(userID, db)
	}

	// Update the count of user notifications, total and unread.
	countOfUnreadNotifications, err := model.GetCountOfUserUnreadNotifications(userID, db)
	if err != nil {
		return err
	}

	user := model.User{}
	user.ID = notification.UserID
	user.UpdateCountOfNotifications(countOfNotifications, countOfUnreadNotifications, db)

	return nil
}
