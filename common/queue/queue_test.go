package queue

import "testing"

// producerStub records the messages it receives.
type producerStub struct{ messages []string }

func (p *producerStub) SendMessageToTopic(message, topic string) {
	p.messages = append(p.messages, topic+":"+message)
}

func TestQueue_send(t *testing.T) {
	producer := &producerStub{}

	New(producer).send("message", "topic")

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

			tt.queue.send("message", "topic")

			if entry := hook.LastEntry(); entry == nil || entry.Data["topic"] != "topic" {
				t.Errorf("the failure must be logged, got %+v", entry)
			}
		})
	}
}
