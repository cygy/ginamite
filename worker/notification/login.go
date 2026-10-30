package notification

import (
	"context"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
)

// NewLogin decodes the payload of a login and notifies the user that a login occurs on their account.
// The user is only notified by email.
func NewLogin(ctx context.Context, payload []byte, getNotificationTargetsByUserAndTypeFunc notifications.GetNotificationTargetsByUserAndTypeFunc) {
	p := queue.UserNotificationNewLogin{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

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
