package queue

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/log"

	"github.com/IBM/sarama/mocks"
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

// capture returns a queue using a Kafka client with a fake producer, and a function that runs send and returns the
// message it sent.
func capture(t *testing.T) (*Queue, func(send func()) sent) {
	t.Helper()

	config := mocks.NewTestConfig()
	config.Producer.Return.Successes = true
	producer := mocks.NewAsyncProducer(t, config)

	client := kafka.NewClient()
	client.HasProducer = true
	client.Producer = producer

	return New(client), func(send func()) sent {
		t.Helper()
		producer.ExpectInputAndSucceed()
		send()

		select {
		case msg := <-producer.Successes():
			value, err := msg.Value.Encode()
			if err != nil {
				t.Fatal(err)
			}

			var m Message
			if err := json.Unmarshal(value, &m); err != nil {
				t.Fatalf("invalid message %q: %v", value, err)
			}

			return sent{topic: msg.Topic, message: m}
		case <-time.After(2 * time.Second):
			t.Fatal("no message sent")
			return sent{}
		}
	}
}
