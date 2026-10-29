package notification

import (
	"context"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
)

// Handler : must return a message type (string) and the associated payload to send to the notification.
type Handler func() (string, interface{})

// Event : settings of an event.
type Event struct {
	UserID                     string
	Type                       string
	EmailFunc                  Handler
	ChromeFunc                 Handler
	SafariFunc                 Handler
	FirefoxFunc                Handler
	IOSFunc                    Handler
	AndroidFunc                Handler
	GetNotificationTargetsFunc notifications.GetNotificationTargetsByUserAndTypeFunc
}

// Dispatch : send the different notifications from the event.
func (event Event) Dispatch(ctx context.Context) {
	if event.GetNotificationTargetsFunc == nil {
		return
	}

	// Get the notification preferences of the user.
	targets, _ := event.GetNotificationTargetsFunc(ctx, event.UserID, event.Type)

	// Send a notification for each target.
	for _, target := range targets {
		switch target.Target {
		case notifications.TargetEmail:
			if event.EmailFunc != nil {
				messageType, payload := event.EmailFunc()
				queue.SendMail(messageType, payload)
			}
		case notifications.TargetChrome:
			if event.ChromeFunc != nil {
				messageType, payload := event.ChromeFunc()
				queue.SendChromeNotification(messageType, payload)
			}
		case notifications.TargetSafari:
			if event.SafariFunc != nil {
				messageType, payload := event.SafariFunc()
				queue.SendSafariNotification(messageType, payload)
			}
		case notifications.TargetFirefox:
			if event.FirefoxFunc != nil {
				messageType, payload := event.FirefoxFunc()
				queue.SendFirefoxNotification(messageType, payload)
			}
		case notifications.TargetIOS:
			if event.IOSFunc != nil {
				messageType, payload := event.IOSFunc()
				queue.SendIOSNotification(messageType, payload)
			}
		case notifications.TargetAndroid:
			if event.AndroidFunc != nil {
				messageType, payload := event.AndroidFunc()
				queue.SendAndroidNotification(messageType, payload)
			}
		default:
		}
	}
}
