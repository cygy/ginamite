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
