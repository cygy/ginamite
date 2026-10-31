package queue

import "testing"

// producerStub records the messages it receives.
type producerStub struct{ messages []string }

func (p *producerStub) SendMessageToTopic(message, topic string) {
	p.messages = append(p.messages, topic+":"+message)
}

func TestQueue_send(t *testing.T) {
	producer := &producerStub{}

	New(producer).send("", "message", "topic")

	if len(producer.messages) != 1 || producer.messages[0] != "topic:message" {
		t.Errorf("messages = %v", producer.messages)
	}
}

func TestQueue_send_NoProducer(t *testing.T) {
	tests := []struct {
		name  string
		queue *Queue
	}{
		{"nil queue", nil},
		{"nil producer", New(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()

			tt.queue.send("", "message", "topic")

			if entry := hook.LastEntry(); entry == nil || entry.Data["topic"] != "topic" {
				t.Errorf("the failure must be logged, got %+v", entry)
			}
		})
	}
}

// keyedProducerStub records the keys of the messages it receives.
type keyedProducerStub struct {
	producerStub
	keys []string
}

func (p *keyedProducerStub) SendKeyedMessageToTopic(key, message, topic string) {
	p.keys = append(p.keys, key)
	p.SendMessageToTopic(message, topic)
}

func TestQueue_send_Keyed(t *testing.T) {
	producer := &keyedProducerStub{}
	q := New(producer)

	q.send("user-1", "a", "topic")
	q.send("", "b", "topic")

	if len(producer.keys) != 1 || producer.keys[0] != "user-1" || len(producer.messages) != 2 {
		t.Errorf("keys = %v, messages = %v", producer.keys, producer.messages)
	}
}

type customKey struct{ ID string }

func (c customKey) PartitionKey() string { return "custom-" + c.ID }

func TestPartitionKey(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		want    string
	}{
		{"user id", TaskDeleteUser{UserID: "u1"}, "u1"},
		{"pointer", &TaskEnableUser{UserID: "u2"}, "u2"},
		{"user id before email", MailNewLogin{EmailAddress: "a@b.c", UserID: "u3"}, "u3"},
		{"email", MailNotification{EmailAddress: "a@b.c"}, "a@b.c"},
		{"derived type", MailRegistrationWelcome{EmailAddress: "d@e.f"}, "d@e.f"},
		{"email if no user id", MailNewLogin{EmailAddress: "a@b.c"}, "a@b.c"},
		{"keyed", customKey{ID: "7"}, "custom-7"},
		{"no key field", TaskIPLocation{IPAddress: "1.2.3.4"}, ""},
		{"not a struct", map[string]string{"user_id": "u"}, ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := partitionKey(tt.payload); got != tt.want {
				t.Errorf("partitionKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
