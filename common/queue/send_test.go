package queue

import (
	"testing"
	"time"
)

func TestSenders(t *testing.T) {
	next := capture(t)

	tests := []struct {
		name      string
		send      func()
		wantTopic string
		wantType  string
	}{
		{"CreateTask", func() { CreateTask("custom", nil) }, TopicTasks, "custom"},
		{"GetIPLocation", func() { GetIPLocation(TaskIPLocation{}) }, TopicTasks, MessageTaskIPLocation},
		{"DeleteUser", func() { DeleteUser(TaskDeleteUser{}) }, TopicTasks, MessageTaskDeleteUser},
		{"DisableUser", func() { DisableUser(TaskDisableUser{}) }, TopicTasks, MessageTaskDisableUser},
		{"EnableUser", func() { EnableUser(TaskEnableUser{}) }, TopicTasks, MessageTaskEnableUser},
		{"UpdateUserSocialNetworks", func() { UpdateUserSocialNetworks(TaskUpdateUserSocialNetworks{}) }, TopicTasks, MessageTaskUpdateUserSocialNetworks},
		{"RegistrationDone", func() { RegistrationDone(TaskRegistrationDone{}) }, TopicTasks, MessageTaskRegistrationDone},
		{"RegistrationValidated", func() { RegistrationValidated(TaskRegistrationValidated{}) }, TopicTasks, MessageTaskRegistrationValidated},
		{"InvalidateCache", func() { InvalidateCache("custom", nil) }, TopicCache, "custom"},
		{"InvalidAuthToken", func() { InvalidAuthToken("id", time.Unix(0, 0)) }, TopicCache, MessageCacheInvalidAuthToken},
		{"InvalidAuthTokens", func() { InvalidAuthTokens(map[string]time.Time{"id": time.Unix(0, 0)}) }, TopicCache, MessageCacheInvalidAuthToken},
		{"NotifyUsersWithAbility", func() { NotifyUsersWithAbility("admin", "custom", nil) }, TopicGroupNotification, "custom"},
		{"NotifyUser", func() { NotifyUser("custom", nil) }, TopicUserNotification, "custom"},
		{"SendMail", func() { SendMail(MessageMailNewLogin, nil) }, TopicMail, MessageMailNewLogin},
		{"SendChromeNotification", func() { SendChromeNotification("custom", nil) }, TopicChrome, "custom"},
		{"SendSafariNotification", func() { SendSafariNotification("custom", nil) }, TopicSafari, "custom"},
		{"SendFirefoxNotification", func() { SendFirefoxNotification("custom", nil) }, TopicFirefox, "custom"},
		{"SendIOSNotification", func() { SendIOSNotification("custom", nil) }, TopicIOS, "custom"},
		{"SendAndroidNotification", func() { SendAndroidNotification("custom", nil) }, TopicAndroid, "custom"},
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
	next := capture(t)
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	got := next(func() { InvalidAuthToken("tok", date) })

	payload := got.message.Payload.(map[string]any)["tokens"].(map[string]any)
	if payload["tok"] != "2026-01-02T03:04:05Z" {
		t.Errorf("payload = %v", got.message.Payload)
	}
}

func TestNotifyUsersWithAbility_Payload(t *testing.T) {
	next := capture(t)

	got := next(func() { NotifyUsersWithAbility("admin", "custom", map[string]string{"k": "v"}) })

	payload := got.message.Payload.(map[string]any)
	if payload["ability"] != "admin" || payload["payload"].(map[string]any)["k"] != "v" {
		t.Errorf("payload = %v", payload)
	}
}
