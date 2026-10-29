package functions

import (
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/gin-gonic/gin"
)

func checkUserIDAndNotificationID(c *gin.Context, userID, notificationID string) (*model.Notification, error) {
	// Check the userID and the notificationID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return nil, err
	}
	if _, err := bson.ObjectIDFromHex(notificationID); err != nil {
		return nil, err
	}

	mongoSession := session.Get(c)

	notification := &model.Notification{}
	if err := notification.GetByID(c, notificationID, mongoSession); err != nil {
		return nil, err
	}

	if notification.UserID.Hex() != userID {
		return nil, errors.NotFound()
	}

	return notification, nil
}
