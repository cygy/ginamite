package worker

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/internal/workertest"
	"github.com/cygy/ginamite/worker/mail"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

// message returns the JSON of a message of Kafka.
func message(t *testing.T, messageType string, payload any) []byte {
	t.Helper()

	b, err := json.Marshal(queue.Message{Type: messageType, Payload: payload})
	if err != nil {
		t.Fatal(err)
	}

	return b
}

// consumer returns the action of the consumer of the topic.
func consumer(t *testing.T, s *Server, topic string) func([]byte) {
	t.Helper()

	for _, c := range s.CreateKafkaConsumers() {
		if c.Topic == topic {
			return c.Action
		}
	}

	t.Fatalf("no consumer for the topic %q", topic)

	return nil
}

// warnings returns the types of the messages logged as unknown, for the topic.
func warnings(topic string) []string {
	types := []string{}
	for _, entry := range workertest.Hook.AllEntries() {
		if entry.Message == "unknown kafka message" && entry.Data["topic"] == topic {
			types = append(types, entry.Data["type"].(string))
		}
	}

	return types
}

func TestServer_CreateKafkaConsumers(t *testing.T) {
	var topics []string
	for _, c := range NewServer().CreateKafkaConsumers() {
		if c.Action == nil {
			t.Errorf("the consumer of %q has no action", c.Topic)
		}
		topics = append(topics, c.Topic)
	}

	want := []string{queue.TopicTasks, queue.TopicGroupNotification, queue.TopicUserNotification, queue.TopicMail}
	if !slices.Equal(topics, want) {
		t.Errorf("topics = %v, want %v", topics, want)
	}
}

func TestServer_CreateKafkaConsumers_Configured(t *testing.T) {
	s := NewServer()
	s.Config.WorkerConsumers = []config.ConsumerConfiguration{
		{Topic: queue.TopicMail, Group: "mails"},
		{Topic: "unknown"},
		{Topic: queue.TopicTasks},
	}

	workertest.Hook.Reset()
	consumers := s.CreateKafkaConsumers()

	if len(consumers) != 2 {
		t.Fatalf("consumers = %+v, want the mails and the tasks", consumers)
	}
	if consumers[0].Topic != queue.TopicMail || consumers[0].Group != "mails" || consumers[0].Action == nil {
		t.Errorf("consumers[0] = %+v", consumers[0])
	}
	if consumers[1].Topic != queue.TopicTasks || consumers[1].Group != "" || consumers[1].Action == nil {
		t.Errorf("consumers[1] = %+v", consumers[1])
	}

	var unknown bool
	for _, entry := range workertest.Hook.AllEntries() {
		if entry.Message == "unknown kafka topic of the worker" && entry.Data["topic"] == "unknown" {
			unknown = true
		}
	}
	if !unknown {
		t.Error("an unknown topic must be logged")
	}
}

func TestServer_CreateKafkaConsumers_WithoutConfiguration(t *testing.T) {
	if consumers := (&Server{}).CreateKafkaConsumers(); len(consumers) != 4 {
		t.Errorf("consumers = %+v, want the 4 topics", consumers)
	}
}

func TestServer_handleTask(t *testing.T) {
	var calls []string
	record := func(name string) func(context.Context, string) {
		return func(_ context.Context, userID string) { calls = append(calls, name+":"+userID) }
	}
	revoke := func(name string) func(context.Context, string) map[string]time.Time {
		return func(_ context.Context, userID string) map[string]time.Time {
			calls = append(calls, name+":"+userID)
			return nil
		}
	}

	s := &Server{Config: config.NewConfiguration(), Functions: Handlers{
		SaveIPAddressDetails: func(_ context.Context, IPAddress, tokenID string, _ func(string) *location.IPAddressDetails) {
			calls = append(calls, "ip:"+IPAddress+":"+tokenID)
		},
		DeleteUserByID:               revoke("delete"),
		DisableUserByID:              revoke("disable"),
		EnableUserByID:               record("enable"),
		UpdateUserSocialNetworksByID: record("social"),
		RegistrationDone:             record("done"),
		RegistrationValidated:        record("validated"),
	}}
	action := consumer(t, s, queue.TopicTasks)

	tests := []struct {
		messageType string
		payload     any
		want        string
	}{
		{queue.MessageTaskIPLocation, queue.TaskIPLocation{IPAddress: "1.2.3.4", TokenID: "t"}, "ip:1.2.3.4:t"},
		{queue.MessageTaskDeleteUser, queue.TaskDeleteUser{UserID: "u"}, "delete:u"},
		{queue.MessageTaskDisableUser, queue.TaskDisableUser{UserID: "u"}, "disable:u"},
		{queue.MessageTaskEnableUser, queue.TaskEnableUser{UserID: "u"}, "enable:u"},
		{queue.MessageTaskUpdateUserSocialNetworks, queue.TaskUpdateUserSocialNetworks{UserID: "u"}, "social:u"},
		{queue.MessageTaskRegistrationDone, queue.TaskRegistrationDone{UserID: "u"}, "done:u"},
		{queue.MessageTaskRegistrationValidated, queue.TaskRegistrationValidated{UserID: "u"}, "validated:u"},
	}

	for _, tt := range tests {
		t.Run(tt.messageType, func(t *testing.T) {
			calls = nil
			action(message(t, tt.messageType, tt.payload))

			if !slices.Equal(calls, []string{tt.want}) {
				t.Errorf("calls = %v, want [%s]", calls, tt.want)
			}
		})
	}

	t.Run("custom message", func(t *testing.T) {
		var got string
		s.Functions.HandleTask = func(_ context.Context, messageType string, payload []byte) bool {
			got = messageType + string(payload)
			return true
		}
		workertest.Hook.Reset()

		action(message(t, "custom", map[string]string{"a": "b"}))

		if got != `custom{"a":"b"}` || len(warnings(queue.TopicTasks)) != 0 {
			t.Errorf("got %q, warnings %v", got, warnings(queue.TopicTasks))
		}
	})

	t.Run("unknown message", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Functions.HandleTask = func(context.Context, string, []byte) bool { return false }
		action(message(t, "unknown", nil))
		s.Functions.HandleTask = nil
		action(message(t, "unknown2", nil))

		if got := warnings(queue.TopicTasks); !slices.Equal(got, []string{"unknown", "unknown2"}) {
			t.Errorf("warnings = %v", got)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		calls = nil
		action([]byte("{"))

		if len(calls) != 0 {
			t.Errorf("calls = %v", calls)
		}
	})

	t.Run("panic recovered", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Functions.EnableUserByID = func(context.Context, string) { panic("boom") }

		action(message(t, queue.MessageTaskEnableUser, queue.TaskEnableUser{UserID: "u"}))

		recovered := false
		for _, entry := range workertest.Hook.AllEntries() {
			recovered = recovered || entry.Data["panic"] == "boom"
		}
		if !recovered {
			t.Error("the panic is not recovered")
		}
	})
}

func TestServer_handleGroupNotification(t *testing.T) {
	s := &Server{}
	action := consumer(t, s, queue.TopicGroupNotification)
	value := message(t, "type", queue.GroupNotification{Ability: "admin", Payload: map[string]string{"a": "b"}})

	t.Run("handled", func(t *testing.T) {
		var got string
		s.Functions.HandleNotificationForUsersWithAbility = func(_ context.Context, ability, messageType string, payload []byte) bool {
			got = ability + ":" + messageType + ":" + string(payload)
			return true
		}
		workertest.Hook.Reset()

		action(value)

		if got != `admin:type:{"a":"b"}` || len(warnings(queue.TopicGroupNotification)) != 0 {
			t.Errorf("got %q, warnings %v", got, warnings(queue.TopicGroupNotification))
		}
	})

	t.Run("not handled", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Functions.HandleNotificationForUsersWithAbility = func(context.Context, string, string, []byte) bool { return false }
		action(value)
		s.Functions.HandleNotificationForUsersWithAbility = nil
		action(value)

		if got := warnings(queue.TopicGroupNotification); !slices.Equal(got, []string{"type", "type"}) {
			t.Errorf("warnings = %v", got)
		}
	})

	t.Run("invalid messages", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Functions.HandleNotificationForUsersWithAbility = func(context.Context, string, string, []byte) bool {
			t.Error("the function must not be called")
			return true
		}

		action([]byte("{"))
		action(message(t, "type", "not a group notification"))
	})
}

func TestServer_handleUserNotification(t *testing.T) {
	var gotTarget string
	s := &Server{Functions: Handlers{
		GetNotificationTargetsByUserAndType: func(_ context.Context, userID, notificationType string) ([]notifications.Target, error) {
			gotTarget = userID + ":" + notificationType
			return nil, nil
		},
	}}
	action := consumer(t, s, queue.TopicUserNotification)

	t.Run("new login", func(t *testing.T) {
		action(message(t, notifications.TypeNewLogin, queue.UserNotificationNewLogin{UserID: "u"}))

		if gotTarget != "u:"+notifications.TypeNewLogin {
			t.Errorf("targets requested for %q", gotTarget)
		}
	})

	t.Run("custom message", func(t *testing.T) {
		var got string
		s.Functions.HandleUserNotification = func(ctx context.Context, messageType string, payload []byte, getTargets notifications.GetNotificationTargetsByUserAndTypeFunc) bool {
			_, _ = getTargets(ctx, "u2", "custom")
			got = messageType + string(payload)
			return true
		}
		workertest.Hook.Reset()

		action(message(t, "custom", map[string]string{"a": "b"}))

		if got != `custom{"a":"b"}` || gotTarget != "u2:custom" || len(warnings(queue.TopicUserNotification)) != 0 {
			t.Errorf("got %q, targets %q, warnings %v", got, gotTarget, warnings(queue.TopicUserNotification))
		}
	})

	t.Run("unknown message", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Functions.HandleUserNotification = func(context.Context, string, []byte, notifications.GetNotificationTargetsByUserAndTypeFunc) bool {
			return false
		}
		action(message(t, "unknown", nil))
		s.Functions.HandleUserNotification = nil
		action(message(t, "unknown2", nil))

		if got := warnings(queue.TopicUserNotification); !slices.Equal(got, []string{"unknown", "unknown2"}) {
			t.Errorf("warnings = %v", got)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		action([]byte("{"))
	})
}

func TestServer_handleMail(t *testing.T) {
	validFunc := func(context.Context, string) (bool, error) { return true, nil }
	s := &Server{Functions: Handlers{IsEmailAddressValid: validFunc}}
	action := consumer(t, s, queue.TopicMail)

	t.Run("built-in message", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Functions.HandleMail = func(context.Context, string, []byte, mail.IsEmailAddressValidFunc) bool {
			t.Error("a built-in message must not be passed to the custom handler")
			return true
		}

		// The payload is invalid: the message is handled without sending any mail.
		action(message(t, queue.MessageMailNewLogin, "invalid payload"))

		if len(warnings(queue.TopicMail)) != 0 {
			t.Errorf("warnings = %v", warnings(queue.TopicMail))
		}
	})

	t.Run("custom message", func(t *testing.T) {
		var got string
		var checked bool
		s.Functions.HandleMail = func(ctx context.Context, messageType string, payload []byte, isValid mail.IsEmailAddressValidFunc) bool {
			checked, _ = isValid(ctx, "john@example.com")
			got = messageType + string(payload)
			return true
		}
		workertest.Hook.Reset()

		action(message(t, "custom", map[string]string{"a": "b"}))

		if got != `custom{"a":"b"}` || !checked || len(warnings(queue.TopicMail)) != 0 {
			t.Errorf("got %q, checked %v, warnings %v", got, checked, warnings(queue.TopicMail))
		}
	})

	t.Run("unknown message", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Functions.HandleMail = func(context.Context, string, []byte, mail.IsEmailAddressValidFunc) bool { return false }
		action(message(t, "unknown", nil))
		s.Functions.HandleMail = nil
		action(message(t, "unknown2", nil))

		if got := warnings(queue.TopicMail); !slices.Equal(got, []string{"unknown", "unknown2"}) {
			t.Errorf("warnings = %v", got)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		action([]byte("{"))
	})
}
