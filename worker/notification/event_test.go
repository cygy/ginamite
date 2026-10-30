package notification

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

// handler returns a handler of a message of the given type.
func handler(messageType string) Handler {
	return func() (string, any) { return messageType, map[string]string{"k": messageType} }
}

// targets returns a function returning the targets.
func targets(targets ...string) notifications.GetNotificationTargetsByUserAndTypeFunc {
	return func(_ context.Context, userID, notificationType string) ([]notifications.Target, error) {
		result := []notifications.Target{}
		for _, target := range targets {
			result = append(result, notifications.Target{Target: target, Value: userID + ":" + notificationType})
		}

		return result, nil
	}
}

func TestEvent_Dispatch(t *testing.T) {
	event := Event{
		UserID:      "u1",
		Type:        "type",
		EmailFunc:   handler("email"),
		ChromeFunc:  handler("chrome"),
		SafariFunc:  handler("safari"),
		FirefoxFunc: handler("firefox"),
		IOSFunc:     handler("ios"),
		AndroidFunc: handler("android"),
	}

	tests := []struct {
		target string
		topic  string
	}{
		{notifications.TargetEmail, queue.TopicMail},
		{notifications.TargetChrome, queue.TopicChrome},
		{notifications.TargetSafari, queue.TopicSafari},
		{notifications.TargetFirefox, queue.TopicFirefox},
		{notifications.TargetIOS, queue.TopicIOS},
		{notifications.TargetAndroid, queue.TopicAndroid},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			q := workertest.CaptureQueue(t)

			e := event
			e.GetNotificationTargetsFunc = targets(tt.target)
			e.Dispatch(context.Background())

			m, ok := q.Next()
			if !ok || m.Topic != tt.topic || m.Type != tt.target || m.Payload["k"] != tt.target {
				t.Errorf("message = %+v, %v", m, ok)
			}
			q.Empty()
		})
	}

	t.Run("all the targets", func(t *testing.T) {
		q := workertest.CaptureQueue(t)

		e := event
		e.GetNotificationTargetsFunc = targets(notifications.TargetEmail, notifications.TargetIOS, notifications.TargetAndroid)
		e.Dispatch(context.Background())

		var got []string
		for {
			m, ok := q.Next()
			if !ok {
				break
			}
			got = append(got, m.Type)
		}

		// The producer is asynchronous: the order of the messages is not guaranteed.
		slices.Sort(got)
		if want := []string{"android", "email", "ios"}; !slices.Equal(got, want) {
			t.Errorf("messages = %v, want %v", got, want)
		}
	})

	t.Run("targets without handler or unknown", func(t *testing.T) {
		q := workertest.CaptureQueue(t)

		e := Event{UserID: "u1", Type: "type", GetNotificationTargetsFunc: targets(notifications.TargetEmail, "unknown")}
		e.Dispatch(context.Background())

		q.Empty()
	})

	t.Run("targets of the user and of the type", func(t *testing.T) {
		var gotUser, gotType string
		e := Event{UserID: "u1", Type: "type", GetNotificationTargetsFunc: func(_ context.Context, userID, notificationType string) ([]notifications.Target, error) {
			gotUser, gotType = userID, notificationType
			return nil, nil
		}}
		e.Dispatch(context.Background())

		if gotUser != "u1" || gotType != "type" {
			t.Errorf("targets requested for %q, %q", gotUser, gotType)
		}
	})

	t.Run("targets error", func(t *testing.T) {
		q := workertest.CaptureQueue(t)
		workertest.Hook.Reset()

		e := event
		e.GetNotificationTargetsFunc = func(context.Context, string, string) ([]notifications.Target, error) {
			return []notifications.Target{{Target: notifications.TargetEmail}}, errors.New("boom")
		}
		e.Dispatch(context.Background())

		q.Empty()
		if entry := workertest.Hook.LastEntry(); entry == nil || entry.Data["error"] != "boom" {
			t.Errorf("log = %+v", entry)
		}
	})

	t.Run("undefined function", func(t *testing.T) {
		Event{UserID: "u1"}.Dispatch(context.Background())
	})
}
