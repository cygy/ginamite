package queue

import (
	"testing"
	"time"
)

func TestSenders(t *testing.T) {
	q, next := capture(t)

	tests := []struct {
		name      string
		send      func()
		wantTopic string
		wantType  string
	}{
		{"CreateTask", func() { q.CreateTask("custom", nil) }, TopicTasks, "custom"},
		{"GetIPLocation", func() { q.GetIPLocation(TaskIPLocation{}) }, TopicTaskLocation, MessageTaskIPLocation},
		{"DeleteUser", func() { q.DeleteUser(TaskDeleteUser{}) }, TopicTaskUser, MessageTaskDeleteUser},
		{"DisableUser", func() { q.DisableUser(TaskDisableUser{}) }, TopicTaskUser, MessageTaskDisableUser},
		{"EnableUser", func() { q.EnableUser(TaskEnableUser{}) }, TopicTaskUser, MessageTaskEnableUser},
		{"UpdateUserSocialNetworks", func() { q.UpdateUserSocialNetworks(TaskUpdateUserSocialNetworks{}) }, TopicTaskUser, MessageTaskUpdateUserSocialNetworks},
		{"RegistrationDone", func() { q.RegistrationDone(TaskRegistrationDone{}) }, TopicTaskRegistration, MessageTaskRegistrationDone},
		{"RegistrationValidated", func() { q.RegistrationValidated(TaskRegistrationValidated{}) }, TopicTaskRegistration, MessageTaskRegistrationValidated},
		{"InvalidateCache", func() { q.InvalidateCache("custom", nil) }, TopicCache, "custom"},
		{"InvalidAuthToken", func() { q.InvalidAuthToken("id", time.Unix(0, 0)) }, TopicCache, MessageCacheInvalidAuthToken},
		{"InvalidAuthTokens", func() { q.InvalidAuthTokens(map[string]time.Time{"id": time.Unix(0, 0)}) }, TopicCache, MessageCacheInvalidAuthToken},
		{"NotifyUsersWithAbility", func() { q.NotifyUsersWithAbility("admin", "custom", nil) }, TopicGroupNotification, "custom"},
		{"NotifyUser", func() { q.NotifyUser("custom", nil) }, TopicUserNotification, "custom"},
		{"SendMail", func() { q.SendMail(MessageMailNewLogin, nil) }, TopicMail, MessageMailNewLogin},
		{"SendChromeNotification", func() { q.SendChromeNotification("custom", nil) }, TopicChrome, "custom"},
		{"SendSafariNotification", func() { q.SendSafariNotification("custom", nil) }, TopicSafari, "custom"},
		{"SendFirefoxNotification", func() { q.SendFirefoxNotification("custom", nil) }, TopicFirefox, "custom"},
		{"SendIOSNotification", func() { q.SendIOSNotification("custom", nil) }, TopicIOS, "custom"},
		{"SendAndroidNotification", func() { q.SendAndroidNotification("custom", nil) }, TopicAndroid, "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := next(tt.send)

			if got.topic != tt.wantTopic || got.message.Type != tt.wantType {
				t.Errorf("sent %q to %q, want %q to %q", got.message.Type, got.topic, tt.wantType, tt.wantTopic)
			}
		})
	}
}

func TestInvalidAuthToken_Payload(t *testing.T) {
	q, next := capture(t)
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	got := next(func() { q.InvalidAuthToken("tok", date) })

	payload := got.message.Payload.(map[string]any)["tokens"].(map[string]any)
	if payload["tok"] != "2026-01-02T03:04:05Z" {
		t.Errorf("payload = %v", got.message.Payload)
	}
}

func TestNotifyUsersWithAbility_Payload(t *testing.T) {
	q, next := capture(t)

	got := next(func() { q.NotifyUsersWithAbility("admin", "custom", map[string]string{"k": "v"}) })

	payload := got.message.Payload.(map[string]any)
	if payload["ability"] != "admin" || payload["payload"].(map[string]any)["k"] != "v" {
		t.Errorf("payload = %v", payload)
	}
}
