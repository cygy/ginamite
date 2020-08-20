package notifications

import (
	"fmt"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
)

// GetNotifications returns the available notifications.
func GetNotifications(supportedNotifications []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := context.GetLocale(c)

		localizedNotifications, ok := notifications[locale]
		if !ok {
			t := localization.Translate(locale)

			localizedNotifications = make([]localizedNotification, len(supportedNotifications))
			for i, notificationType := range supportedNotifications {
				localizedNotifications[i] = localizedNotification{
					Type:        notificationType,
					Name:        t(fmt.Sprintf("notification.%s.name", notificationType)),
					Description: t(fmt.Sprintf("notification.%s.description", notificationType)),
				}
			}

			notifications[locale] = localizedNotifications
		}

		response.Ok(c, response.H{
			"notifications": localizedNotifications,
		})
	}

}
