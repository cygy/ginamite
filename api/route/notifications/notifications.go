// Package notifications provides the handler listing the supported notifications.
package notifications

import (
	"fmt"
	"sync"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/response"

	"github.com/gin-gonic/gin"
)

// GetNotifications returns the handler listing the supported notifications. The localized notifications are cached
// by locale in the handler.
func GetNotifications(supportedNotifications []string) gin.HandlerFunc {
	var mutex sync.Mutex
	notifications := map[string][]localizedNotification{} // Locale > []struct

	return func(c *gin.Context) {
		locale := ctxLocale.Get(c)

		mutex.Lock()
		defer mutex.Unlock()

		localizedNotifications, ok := notifications[locale]
		if !ok {
			t := ctxkey.Localization.Get(c).Translate(locale)

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
