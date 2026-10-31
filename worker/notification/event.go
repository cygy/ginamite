// Package notification dispatches the notifications of the events to the targets chosen by the users.
package notification

import (
	"context"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
)

// Handler returns the type of a message and its payload.
type Handler func() (string, any)

// Event is an event to notify to a user, with the message to send to each kind of target.
// A user is not notified through the targets without handler.
type Event struct {
	Queue                      *queue.Queue // The queue sending the messages to the targets.
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

// Dispatch sends a message to each target of the user for the type of the event.
// Nothing is sent if the targets can not be retrieved, and the error is logged.
func (event Event) Dispatch(ctx context.Context) {
	if event.GetNotificationTargetsFunc == nil {
		return
	}

	targets, err := event.GetNotificationTargetsFunc(ctx, event.UserID, event.Type)
	if err != nil {
		log.WithField("error", err.Error()).Error("unable to get the notification targets of the user")
		return
	}

	senders := map[string]struct {
		handler Handler
		send    func(messageType string, payload any)
	}{
		notifications.TargetEmail:   {event.EmailFunc, event.Queue.SendMail},
		notifications.TargetChrome:  {event.ChromeFunc, event.Queue.SendChromeNotification},
		notifications.TargetSafari:  {event.SafariFunc, event.Queue.SendSafariNotification},
		notifications.TargetFirefox: {event.FirefoxFunc, event.Queue.SendFirefoxNotification},
		notifications.TargetIOS:     {event.IOSFunc, event.Queue.SendIOSNotification},
		notifications.TargetAndroid: {event.AndroidFunc, event.Queue.SendAndroidNotification},
	}

	for _, target := range targets {
		if sender, ok := senders[target.Target]; ok && sender.handler != nil {
			messageType, payload := sender.handler()
			sender.send(messageType, payload)
		}
	}
}
