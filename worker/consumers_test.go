package worker

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/broker"
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

// consumerErr returns the action of the consumer of the topic.
func consumerErr(t *testing.T, s *Server, topic string) func([]byte) error {
	t.Helper()

	for _, c := range s.CreateConsumers() {
		if c.Topic == topic {
			return c.Action
		}
	}

	t.Fatalf("no consumer for the topic %q", topic)

	return nil
}

// consumer returns the action of the consumer of the topic, which fails the test if the message is not handled.
func consumer(t *testing.T, s *Server, topic string) func([]byte) {
	t.Helper()

	action := consumerErr(t, s, topic)

	return func(value []byte) {
		t.Helper()

		if err := action(value); err != nil {
			t.Errorf("the message %s failed: %v", value, err)
		}
	}
}

// checkUndecodable checks that the consumer of the topic dead-letters a message it can not decode, without a retry.
func checkUndecodable(t *testing.T, s *Server, topic string) {
	t.Helper()

	if err := consumerErr(t, s, topic)([]byte("{")); !broker.IsPermanent(err) {
		t.Errorf("error = %v, want a permanent error", err)
	}
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

func TestServer_CreateConsumers(t *testing.T) {
	var topics []string
	for _, c := range NewServer().CreateConsumers() {
		if c.Action == nil {
			t.Errorf("the consumer of %q has no action", c.Topic)
		}
		topics = append(topics, c.Topic)
	}

	want := []string{
		queue.TopicTasks, queue.TopicTaskLocation, queue.TopicTaskUser, queue.TopicTaskRegistration,
		queue.TopicGroupNotification, queue.TopicUserNotification, queue.TopicMail,
	}
	if !slices.Equal(topics, want) {
		t.Errorf("topics = %v, want %v", topics, want)
	}
}

func TestServer_CreateConsumers_Configured(t *testing.T) {
	s := NewServer()
	s.Config.WorkerConsumers = []config.ConsumerConfiguration{
		{Topic: queue.TopicMail, Group: "mails"},
		{Topic: "unknown"},
		{Topic: queue.TopicTasks},
	}

	workertest.Hook.Reset()
	consumers := s.CreateConsumers()

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

func TestServer_CreateConsumers_WithoutConfiguration(t *testing.T) {
	if consumers := (&Server{}).CreateConsumers(); len(consumers) != 7 {
		t.Errorf("consumers = %+v, want the 7 topics", consumers)
	}
}

func TestServer_handleTask(t *testing.T) {
	var calls []string
	record := func(name string) func(context.Context, string) error {
		return func(_ context.Context, userID string) error {
			calls = append(calls, name+":"+userID)
			return nil
		}
	}
	revoke := func(name string) func(context.Context, string) (map[string]time.Time, error) {
		return func(_ context.Context, userID string) (map[string]time.Time, error) {
			calls = append(calls, name+":"+userID)
			return nil, nil
		}
	}

	s := &Server{Config: config.NewConfiguration(), Storage: Storage{
		Locations: fakeLocations{save: func(_ context.Context, IPAddress, tokenID string, _ location.DetailsFunc) error {
			calls = append(calls, "ip:"+IPAddress+":"+tokenID)
			return nil
		}},
		Users: fakeUsers{del: revoke("delete"), disable: revoke("disable"), enable: record("enable"), social: record("social")},
	}, Hooks: Hooks{
		RegistrationDone:      record("done"),
		RegistrationValidated: record("validated"),
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

	// The built-in tasks are handled by the topic of their type, and by the legacy topic "task".
	for _, tt := range tests {
		topic := map[string]string{
			queue.MessageTaskIPLocation:               queue.TopicTaskLocation,
			queue.MessageTaskDeleteUser:               queue.TopicTaskUser,
			queue.MessageTaskDisableUser:              queue.TopicTaskUser,
			queue.MessageTaskEnableUser:               queue.TopicTaskUser,
			queue.MessageTaskUpdateUserSocialNetworks: queue.TopicTaskUser,
			queue.MessageTaskRegistrationDone:         queue.TopicTaskRegistration,
			queue.MessageTaskRegistrationValidated:    queue.TopicTaskRegistration,
		}[tt.messageType]

		t.Run(tt.messageType+" on "+topic, func(t *testing.T) {
			calls = nil
			consumer(t, s, topic)(message(t, tt.messageType, tt.payload))

			if !slices.Equal(calls, []string{tt.want}) {
				t.Errorf("calls = %v, want [%s]", calls, tt.want)
			}
		})
	}

	t.Run("custom message", func(t *testing.T) {
		var got string
		s.Hooks.HandleTask = func(_ context.Context, messageType string, payload []byte) (bool, error) {
			got = messageType + string(payload)
			return true, nil
		}
		workertest.Hook.Reset()

		action(message(t, "custom", map[string]string{"a": "b"}))

		if got != `custom{"a":"b"}` || len(warnings(queue.TopicTasks)) != 0 {
			t.Errorf("got %q, warnings %v", got, warnings(queue.TopicTasks))
		}
	})

	t.Run("unknown message", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Hooks.HandleTask = func(context.Context, string, []byte) (bool, error) { return false, nil }
		action(message(t, "unknown", nil))
		s.Hooks.HandleTask = nil
		action(message(t, "unknown2", nil))

		if got := warnings(queue.TopicTasks); !slices.Equal(got, []string{"unknown", "unknown2"}) {
			t.Errorf("warnings = %v", got)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		calls = nil
		checkUndecodable(t, s, queue.TopicTasks)

		if len(calls) != 0 {
			t.Errorf("calls = %v", calls)
		}
	})
}

func TestServer_handleTask_Errors(t *testing.T) {
	boom := errors.New("boom")
	fail := func(context.Context, string) error { return boom }

	tests := []struct {
		name        string
		messageType string
		payload     any
		storage     Storage
		hooks       Hooks
	}{
		{"IP location", queue.MessageTaskIPLocation, queue.TaskIPLocation{IPAddress: "1.2.3.4"}, Storage{Locations: fakeLocations{save: func(context.Context, string, string, location.DetailsFunc) error { return boom }}}, Hooks{}},
		{"delete user", queue.MessageTaskDeleteUser, queue.TaskDeleteUser{UserID: "u"}, Storage{Users: fakeUsers{del: func(context.Context, string) (map[string]time.Time, error) { return nil, boom }}}, Hooks{}},
		{"disable user", queue.MessageTaskDisableUser, queue.TaskDisableUser{UserID: "u"}, Storage{Users: fakeUsers{disable: func(context.Context, string) (map[string]time.Time, error) { return nil, boom }}}, Hooks{}},
		{"enable user", queue.MessageTaskEnableUser, queue.TaskEnableUser{UserID: "u"}, Storage{Users: fakeUsers{enable: fail}}, Hooks{}},
		{"social networks", queue.MessageTaskUpdateUserSocialNetworks, queue.TaskUpdateUserSocialNetworks{UserID: "u"}, Storage{Users: fakeUsers{social: fail}}, Hooks{}},
		{"registration done", queue.MessageTaskRegistrationDone, queue.TaskRegistrationDone{UserID: "u"}, Storage{}, Hooks{RegistrationDone: fail}},
		{"registration validated", queue.MessageTaskRegistrationValidated, queue.TaskRegistrationValidated{UserID: "u"}, Storage{}, Hooks{RegistrationValidated: fail}},
		{"custom", "custom", nil, Storage{}, Hooks{HandleTask: func(context.Context, string, []byte) (bool, error) { return true, boom }}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{Config: config.NewConfiguration(), Storage: tt.storage, Hooks: tt.hooks, Queue: workertest.CaptureQueue(t).Sender}

			err := consumerErr(t, s, queue.TopicTasks)(message(t, tt.messageType, tt.payload))

			if !errors.Is(err, boom) || broker.IsPermanent(err) {
				t.Errorf("error = %v, want the error of the function, which a retry can fix", err)
			}
		})
	}

	t.Run("an unhandled custom message is not a failure", func(t *testing.T) {
		workertest.Hook.Reset()
		s := &Server{Hooks: Hooks{HandleTask: func(context.Context, string, []byte) (bool, error) { return false, boom }}}

		if err := consumerErr(t, s, queue.TopicTasks)(message(t, "custom", nil)); err != nil {
			t.Errorf("error = %v", err)
		}
		if got := warnings(queue.TopicTasks); !slices.Equal(got, []string{"custom"}) {
			t.Errorf("warnings = %v", got)
		}
	})
}

func TestServer_handleOtherTopics_Errors(t *testing.T) {
	boom := errors.New("boom")
	valid := func(context.Context, string) (bool, error) { return true, nil }

	t.Run("group notification", func(t *testing.T) {
		s := &Server{Hooks: Hooks{HandleNotificationForUsersWithAbility: func(context.Context, string, string, []byte) (bool, error) { return true, boom }}}

		value := message(t, "type", queue.GroupNotification{Ability: "admin", Payload: map[string]string{"a": "b"}})
		if err := consumerErr(t, s, queue.TopicGroupNotification)(value); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("user notification", func(t *testing.T) {
		s := &Server{Hooks: Hooks{HandleUserNotification: func(context.Context, string, []byte, notifications.GetNotificationTargetsByUserAndTypeFunc) (bool, error) {
			return true, boom
		}}}

		if err := consumerErr(t, s, queue.TopicUserNotification)(message(t, "custom", nil)); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("new login", func(t *testing.T) {
		s := &Server{Storage: Storage{Notifications: fakeNotifications{targets: func(context.Context, string, string) ([]notifications.Target, error) {
			return nil, boom
		}}}}

		value := message(t, notifications.TypeNewLogin, queue.UserNotificationNewLogin{UserID: "u"})
		if err := consumerErr(t, s, queue.TopicUserNotification)(value); !errors.Is(err, boom) || broker.IsPermanent(err) {
			t.Errorf("error = %v, want the error of the targets, which a retry can fix", err)
		}
	})

	t.Run("mail", func(t *testing.T) {
		s := &Server{Storage: Storage{Emails: fakeEmails{isValid: valid}}, Hooks: Hooks{HandleMail: func(context.Context, string, []byte, mail.IsEmailAddressValidFunc) (bool, error) {
			return true, boom
		}}}

		if err := consumerErr(t, s, queue.TopicMail)(message(t, "custom", nil)); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}
	})
}

func TestServer_handleGroupNotification(t *testing.T) {
	s := &Server{}
	action := consumer(t, s, queue.TopicGroupNotification)
	value := message(t, "type", queue.GroupNotification{Ability: "admin", Payload: map[string]string{"a": "b"}})

	t.Run("handled", func(t *testing.T) {
		var got string
		s.Hooks.HandleNotificationForUsersWithAbility = func(_ context.Context, ability, messageType string, payload []byte) (bool, error) {
			got = ability + ":" + messageType + ":" + string(payload)
			return true, nil
		}
		workertest.Hook.Reset()

		action(value)

		if got != `admin:type:{"a":"b"}` || len(warnings(queue.TopicGroupNotification)) != 0 {
			t.Errorf("got %q, warnings %v", got, warnings(queue.TopicGroupNotification))
		}
	})

	t.Run("not handled", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Hooks.HandleNotificationForUsersWithAbility = func(context.Context, string, string, []byte) (bool, error) { return false, nil }
		action(value)
		s.Hooks.HandleNotificationForUsersWithAbility = nil
		action(value)

		if got := warnings(queue.TopicGroupNotification); !slices.Equal(got, []string{"type", "type"}) {
			t.Errorf("warnings = %v", got)
		}
	})

	t.Run("invalid messages", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Hooks.HandleNotificationForUsersWithAbility = func(context.Context, string, string, []byte) (bool, error) {
			t.Error("the function must not be called")
			return true, nil
		}

		checkUndecodable(t, s, queue.TopicGroupNotification)
		if err := consumerErr(t, s, queue.TopicGroupNotification)(message(t, "type", "not a group notification")); !broker.IsPermanent(err) {
			t.Errorf("error = %v, want a permanent error", err)
		}
	})
}

func TestServer_handleUserNotification(t *testing.T) {
	var gotTarget string
	s := &Server{Storage: Storage{Notifications: fakeNotifications{targets: func(_ context.Context, userID, notificationType string) ([]notifications.Target, error) {
		gotTarget = userID + ":" + notificationType
		return nil, nil
	}}}}
	action := consumer(t, s, queue.TopicUserNotification)

	t.Run("new login", func(t *testing.T) {
		action(message(t, notifications.TypeNewLogin, queue.UserNotificationNewLogin{UserID: "u"}))

		if gotTarget != "u:"+notifications.TypeNewLogin {
			t.Errorf("targets requested for %q", gotTarget)
		}
	})

	t.Run("custom message", func(t *testing.T) {
		var got string
		s.Hooks.HandleUserNotification = func(ctx context.Context, messageType string, payload []byte, getTargets notifications.GetNotificationTargetsByUserAndTypeFunc) (bool, error) {
			_, _ = getTargets(ctx, "u2", "custom")
			got = messageType + string(payload)
			return true, nil
		}
		workertest.Hook.Reset()

		action(message(t, "custom", map[string]string{"a": "b"}))

		if got != `custom{"a":"b"}` || gotTarget != "u2:custom" || len(warnings(queue.TopicUserNotification)) != 0 {
			t.Errorf("got %q, targets %q, warnings %v", got, gotTarget, warnings(queue.TopicUserNotification))
		}
	})

	t.Run("unknown message", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Hooks.HandleUserNotification = func(context.Context, string, []byte, notifications.GetNotificationTargetsByUserAndTypeFunc) (bool, error) {
			return false, nil
		}
		action(message(t, "unknown", nil))
		s.Hooks.HandleUserNotification = nil
		action(message(t, "unknown2", nil))

		if got := warnings(queue.TopicUserNotification); !slices.Equal(got, []string{"unknown", "unknown2"}) {
			t.Errorf("warnings = %v", got)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		checkUndecodable(t, s, queue.TopicUserNotification)
	})
}

func TestServer_handleMail(t *testing.T) {
	validFunc := func(context.Context, string) (bool, error) { return true, nil }
	s := &Server{Storage: Storage{Emails: fakeEmails{isValid: validFunc}}}
	action := consumer(t, s, queue.TopicMail)

	t.Run("built-in message", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Hooks.HandleMail = func(context.Context, string, []byte, mail.IsEmailAddressValidFunc) (bool, error) {
			t.Error("a built-in message must not be passed to the custom handler")
			return true, nil
		}

		// The payload is invalid: the message is handled without sending any mail, and dead-lettered.
		if err := consumerErr(t, s, queue.TopicMail)(message(t, queue.MessageMailNewLogin, "invalid payload")); !broker.IsPermanent(err) {
			t.Errorf("error = %v, want a permanent error", err)
		}

		if len(warnings(queue.TopicMail)) != 0 {
			t.Errorf("warnings = %v", warnings(queue.TopicMail))
		}
	})

	t.Run("custom message", func(t *testing.T) {
		var got string
		var checked bool
		s.Hooks.HandleMail = func(ctx context.Context, messageType string, payload []byte, isValid mail.IsEmailAddressValidFunc) (bool, error) {
			checked, _ = isValid(ctx, "john@example.com")
			got = messageType + string(payload)
			return true, nil
		}
		workertest.Hook.Reset()

		action(message(t, "custom", map[string]string{"a": "b"}))

		if got != `custom{"a":"b"}` || !checked || len(warnings(queue.TopicMail)) != 0 {
			t.Errorf("got %q, checked %v, warnings %v", got, checked, warnings(queue.TopicMail))
		}
	})

	t.Run("unknown message", func(t *testing.T) {
		workertest.Hook.Reset()
		s.Hooks.HandleMail = func(context.Context, string, []byte, mail.IsEmailAddressValidFunc) (bool, error) { return false, nil }
		action(message(t, "unknown", nil))
		s.Hooks.HandleMail = nil
		action(message(t, "unknown2", nil))

		if got := warnings(queue.TopicMail); !slices.Equal(got, []string{"unknown", "unknown2"}) {
			t.Errorf("warnings = %v", got)
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		checkUndecodable(t, s, queue.TopicMail)
	})
}
