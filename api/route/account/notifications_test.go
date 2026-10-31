package account

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/notifications"

	"github.com/gin-gonic/gin"
)

func TestGetNotifications(t *testing.T) {
	list := func(err error) func() {
		return func() {
			testFunctions.NotificationsReceivedByUser = func(*gin.Context, string) (any, error) { return []string{"n1"}, err }
		}
	}

	runCases(t, GetNotifications, []testCase{
		{name: "function not defined", userID: testUserID, want: http.StatusInternalServerError},
		{name: "error", userID: testUserID, setup: list(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, setup: list(nil), want: http.StatusOK, check: contains("n1")},
	})
}

func TestGetNotification(t *testing.T) {
	get := func(err error) func() {
		return func() {
			testFunctions.NotificationReceivedByUser = func(*gin.Context, string, string) (any, error) { return "n1", err }
		}
	}

	runCases(t, GetNotification, []testCase{
		{name: "function not defined", userID: testUserID, id: validNotificationID, want: http.StatusInternalServerError},
		{name: "missing ID", userID: testUserID, setup: get(nil), want: http.StatusBadRequest},
		{name: "invalid ID", userID: testUserID, id: "nope", setup: get(nil), want: http.StatusBadRequest},
		{name: "error", userID: testUserID, id: validNotificationID, setup: get(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, id: validNotificationID, setup: get(nil), want: http.StatusOK, check: contains("n1")},
	})
}

// notificationUpdateCases returns the scenarios shared by the handlers updating or deleting a single notification.
func notificationUpdateCases(inject func(err error) func()) []testCase {
	return []testCase{
		{name: "function not defined", userID: testUserID, id: validNotificationID, want: http.StatusInternalServerError},
		{name: "missing ID", userID: testUserID, setup: inject(nil), want: http.StatusBadRequest},
		{name: "invalid ID", userID: testUserID, id: "nope", setup: inject(nil), want: http.StatusBadRequest},
		{name: "error", userID: testUserID, id: validNotificationID, setup: inject(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, id: validNotificationID, setup: inject(nil), want: http.StatusOK},
	}
}

func TestUpdateNotificationRead(t *testing.T) {
	runCases(t, UpdateNotificationRead, notificationUpdateCases(func(err error) func() {
		return func() {
			testFunctions.UpdateUserNotificationRead = func(*gin.Context, string, string) error { return err }
		}
	}))
}

func TestUpdateNotificationNotified(t *testing.T) {
	runCases(t, UpdateNotificationNotified, notificationUpdateCases(func(err error) func() {
		return func() {
			testFunctions.UpdateUserNotificationNotified = func(*gin.Context, string, string) error { return err }
		}
	}))
}

func TestDeleteNotification(t *testing.T) {
	runCases(t, DeleteNotification, notificationUpdateCases(func(err error) func() {
		return func() { testFunctions.DeleteUserNotification = func(*gin.Context, string, string) error { return err } }
	}))
}

func TestGetNotificationsSettings(t *testing.T) {
	runCases(t, GetNotificationsSettings, []testCase{
		{name: "function not defined", userID: testUserID, want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, want: http.StatusOK, check: contains("n1"), setup: func() {
			testFunctions.NotificationsSubscribedByUser = func(*gin.Context, string) []string { return []string{"n1"} }
		}},
	})
}

func TestSaveNotificationsSettings(t *testing.T) {
	body := `{"notifications":["` + notifications.TypeNewLogin + `"]}`
	save := func(err error) func() {
		return func() {
			testFunctions.SubscribeNotificationsToUser = func(*gin.Context, string, []string) error { return err }
		}
	}

	runCases(t, SaveNotificationsSettings, []testCase{
		{name: "function not defined", userID: testUserID, body: body, want: http.StatusInternalServerError},
		{name: "unsupported notification", userID: testUserID, body: `{"notifications":["unknown"]}`, setup: save(nil), want: http.StatusBadRequest},
		{name: "error", userID: testUserID, body: body, setup: save(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: save(nil), want: http.StatusOK},
	})
}

func TestUnsubscribeNotificationByKey(t *testing.T) {
	unsubscribe := func(err error) func() {
		return func() {
			testFunctions.UnsubscribeNotificationFromUser = func(*gin.Context, string, string, string) error { return err }
		}
	}
	validKey := "user." + notifications.TypeNewLogin + ".signature"

	runCases(t, UnsubscribeNotificationByKey, []testCase{
		{name: "function not defined", id: "u", key: validKey, want: http.StatusBadRequest},
		{name: "missing key", setup: unsubscribe(nil), want: http.StatusBadRequest},
		{name: "malformed key", id: "u", key: "user.signature", setup: unsubscribe(nil), want: http.StatusBadRequest},
		{name: "unsupported notification", id: "u", key: "user.unknown.signature", setup: unsubscribe(nil), want: http.StatusBadRequest},
		{name: "error", id: "u", key: validKey, setup: unsubscribe(errBoom), want: http.StatusBadRequest},
		{name: "success", id: "u", key: validKey, setup: unsubscribe(nil), want: http.StatusOK},
	})
}
