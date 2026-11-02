package account

import (
	"context"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/notifications"
)

func TestGetNotifications(t *testing.T) {
	list := func(err error) func() {
		return func() {
			testNotifications.receivedFn = func(context.Context, string, string) (any, error) { return []string{"n1"}, err }
		}
	}

	runCases(t, GetNotifications, []testCase{
		{name: "storage not defined", setup: withoutNotifications, userID: testUserID, want: http.StatusInternalServerError},
		{name: "error", userID: testUserID, setup: list(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, setup: list(nil), want: http.StatusOK, check: contains("n1")},
	})
}

func TestGetNotification(t *testing.T) {
	get := func(err error) func() {
		return func() {
			testNotifications.receivedOneFn = func(context.Context, string, string, string) (any, error) { return "n1", err }
		}
	}

	runCases(t, GetNotification, []testCase{
		{name: "storage not defined", setup: withoutNotifications, userID: testUserID, id: validNotificationID, want: http.StatusInternalServerError},
		{name: "missing ID", userID: testUserID, setup: get(nil), want: http.StatusBadRequest},
		{name: "invalid ID", userID: testUserID, id: "nope", setup: get(nil), want: http.StatusBadRequest},
		{name: "error", userID: testUserID, id: validNotificationID, setup: get(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, id: validNotificationID, setup: get(nil), want: http.StatusOK, check: contains("n1")},
	})
}

// notificationUpdateCases returns the scenarios shared by the handlers updating or deleting a single notification.
func notificationUpdateCases(inject func(err error) func()) []testCase {
	return []testCase{
		{name: "storage not defined", setup: withoutNotifications, userID: testUserID, id: validNotificationID, want: http.StatusInternalServerError},
		{name: "missing ID", userID: testUserID, setup: inject(nil), want: http.StatusBadRequest},
		{name: "invalid ID", userID: testUserID, id: "nope", setup: inject(nil), want: http.StatusBadRequest},
		{name: "error", userID: testUserID, id: validNotificationID, setup: inject(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, id: validNotificationID, setup: inject(nil), want: http.StatusOK},
	}
}

func TestUpdateNotificationRead(t *testing.T) {
	runCases(t, UpdateNotificationRead, notificationUpdateCases(func(err error) func() {
		return func() {
			testNotifications.readFn = func(context.Context, string, string) error { return err }
		}
	}))
}

func TestUpdateNotificationNotified(t *testing.T) {
	runCases(t, UpdateNotificationNotified, notificationUpdateCases(func(err error) func() {
		return func() {
			testNotifications.notifiedFn = func(context.Context, string, string) error { return err }
		}
	}))
}

func TestDeleteNotification(t *testing.T) {
	runCases(t, DeleteNotification, notificationUpdateCases(func(err error) func() {
		return func() { testNotifications.deleteFn = func(context.Context, string, string) error { return err } }
	}))
}

func TestGetNotificationsSettings(t *testing.T) {
	runCases(t, GetNotificationsSettings, []testCase{
		{name: "storage not defined", setup: withoutNotifications, userID: testUserID, want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, want: http.StatusOK, check: contains("n1"), setup: func() {
			testNotifications.subscribedFn = func(context.Context, string) ([]string, error) { return []string{"n1"}, nil }
		}},
	})
}

func TestSaveNotificationsSettings(t *testing.T) {
	body := `{"notifications":["` + notifications.TypeNewLogin + `"]}`
	save := func(err error) func() {
		return func() {
			testNotifications.subscribeFn = func(context.Context, string, []string) error { return err }
		}
	}

	runCases(t, SaveNotificationsSettings, []testCase{
		{name: "storage not defined", setup: withoutNotifications, userID: testUserID, body: body, want: http.StatusInternalServerError},
		{name: "unsupported notification", userID: testUserID, body: `{"notifications":["unknown"]}`, setup: save(nil), want: http.StatusBadRequest},
		{name: "error", userID: testUserID, body: body, setup: save(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: save(nil), want: http.StatusOK},
	})
}

func TestUnsubscribeNotificationByKey(t *testing.T) {
	unsubscribe := func(err error) func() {
		return func() {
			testNotifications.unsubscribeFn = func(context.Context, string, string, string) error { return err }
		}
	}
	validKey := "user." + notifications.TypeNewLogin + ".signature"

	runCases(t, UnsubscribeNotificationByKey, []testCase{
		{name: "storage not defined", setup: withoutNotifications, id: "u", key: validKey, want: http.StatusBadRequest},
		{name: "missing key", setup: unsubscribe(nil), want: http.StatusBadRequest},
		{name: "malformed key", id: "u", key: "user.signature", setup: unsubscribe(nil), want: http.StatusBadRequest},
		{name: "unsupported notification", id: "u", key: "user.unknown.signature", setup: unsubscribe(nil), want: http.StatusBadRequest},
		{name: "error", id: "u", key: validKey, setup: unsubscribe(errBoom), want: http.StatusBadRequest},
		{name: "success", id: "u", key: validKey, setup: unsubscribe(nil), want: http.StatusOK},
	})
}
