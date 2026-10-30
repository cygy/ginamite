package notification

import (
	"context"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
)

// NewLogin : notifies a user that a login occurs on its account.
func NewLogin(ctx context.Context, payload []byte, getNotificationTargetsByUserAndTypeFunc notifications.GetNotificationTargetsByUserAndTypeFunc) {
	p := queue.UserNotificationNewLogin{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	// Dispatch this event to the different notifications.
	event := Event{
		UserID:                     p.UserID,
		Type:                       notifications.TypeNewLogin,
		GetNotificationTargetsFunc: getNotificationTargetsByUserAndTypeFunc,
		EmailFunc: func() (string, any) {
			return queue.MessageMailNewLogin, p
		},
	}
	event.Dispatch(ctx)
}
