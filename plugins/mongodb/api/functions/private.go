package functions

import (
	"errors"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

func checkUserIDAndNotificationID(c *gin.Context, userID, notificationID string) (*model.Notification, error) {
	// Check the userID and the notificationID.
	if !bson.IsObjectIdHex(userID) || !bson.IsObjectIdHex(notificationID) {
		return nil, errors.New("wrong object ID")
	}

	mongoSession := session.Get(c)

	notification := &model.Notification{}
	if err := notification.GetByID(notificationID, mongoSession); err != nil {
		return nil, err
	}

	if notification.UserID.Hex() != userID {
		return nil, commonErrors.NotFound()
	}

	return notification, nil
}
