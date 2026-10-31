package queue

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// hook collects the entries of the logger.
var hook *logtest.Hook

func TestMain(m *testing.M) {
	var logger *logrus.Logger
	logger, hook = logtest.NewNullLogger()
	log.Main = logger.WithField("server", "test")

	os.Exit(m.Run())
}

// sent is a message captured by the fake producer.
type sent struct {
	topic   string
	message Message
}

// capture returns a queue using a Memory broker, and a function that runs send and returns the message it sent.
func capture(t *testing.T) (*Queue, func(send func()) sent) {
	t.Helper()

	memory := broker.NewMemory()

	return New(memory), func(send func()) sent {
		t.Helper()

		before := len(memory.Sent())
		send()

		messages := memory.Sent()
		if len(messages) != before+1 {
			t.Fatalf("%d messages sent, want 1", len(messages)-before)
		}

		var m Message
		if err := json.Unmarshal([]byte(messages[before].Message), &m); err != nil {
			t.Fatalf("invalid message %q: %v", messages[before].Message, err)
		}

		return sent{topic: messages[before].Topic, message: m}
	}
}
