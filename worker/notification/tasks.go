package notification

import (
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
)

// NewLogin : notifies a user that a login occurs on its account.
func NewLogin(payload []byte, getNotificationTargetsByUserAndTypeFunc notifications.GetNotificationTargetsByUserAndTypeFunc) {
	p := queue.UserNotificationNewLogin{}
	queue.UnmarshalPayload(payload, &p)

	// Dispatch this event to the different notifications.
	event := Event{
		UserID:                     p.UserID,
		Type:                       notifications.TypeNewLogin,
		GetNotificationTargetsFunc: getNotificationTargetsByUserAndTypeFunc,
		EmailFunc: func() (string, interface{}) {
			return queue.MessageMailNewLogin, p
		},
	}
	event.Dispatch()
}
