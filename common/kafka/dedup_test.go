package kafka

import (
	"testing"

	"github.com/IBM/sarama"
)

func TestMessageID(t *testing.T) {
	with := &sarama.ConsumerMessage{Headers: []*sarama.RecordHeader{
		nil,
		{Key: []byte("other"), Value: []byte("x")},
		{Key: []byte(HeaderMessageID), Value: []byte("abc")},
	}}
	if got := messageID(with); got != "abc" {
		t.Errorf("messageID() = %q", got)
	}
	if got := messageID(&sarama.ConsumerMessage{}); got != "" {
		t.Errorf("messageID() without header = %q", got)
	}
}

func TestMemoryDeduplicator(t *testing.T) {
	d := NewMemoryDeduplicator(2)

	if d.Handled("g", "1") {
		t.Fatal("unknown message reported as handled")
	}

	d.MarkHandled("g", "1")
	d.MarkHandled("g", "1") // Marking twice does not use a second slot.
	d.MarkHandled("g", "2")
	if !d.Handled("g", "1") || !d.Handled("g", "2") {
		t.Error("the messages are not remembered")
	}
	if d.Handled("other", "1") {
		t.Error("the groups must not share their messages")
	}

	d.MarkHandled("g", "3") // Forgets the oldest one.
	if d.Handled("g", "1") || !d.Handled("g", "2") || !d.Handled("g", "3") {
		t.Error("the oldest message must be forgotten first")
	}

	if NewMemoryDeduplicator(0) == nil {
		t.Error("a size below 1 must be accepted")
	}
}
