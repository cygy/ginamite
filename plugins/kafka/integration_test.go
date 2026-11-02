//go:build integration

package kafka

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/broker"

	"github.com/google/uuid"
)

// TestClient_Integration_SendAndConsume connects to a real Kafka broker, reachable at $KAFKA_HOST:$KAFKA_PORT
// (localhost:9092 if unset), and checks that a message sent to a topic is consumed back. The topic must allow
// auto-creation, the default of a broker. Run it with "go test -tags=integration ./plugins/kafka/...", or
// "make integration".
func TestClient_Integration_SendAndConsume(t *testing.T) {
	host := envOr(t, "KAFKA_HOST", "localhost")
	port := envIntOr(t, "KAFKA_PORT", 9092)
	topic := "integration_" + uuid.NewString()
	message := uuid.NewString()

	received := make(chan string, 1)
	client := NewClient()
	client.Initialize(host, port, "", true, []broker.MessageConsumer{{
		Topic: topic,
		Action: func(value []byte) error {
			received <- string(value)
			return nil
		},
	}})
	t.Cleanup(client.Close)

	// The consumer group starts from the newest offset, and both the producer and the consumer connect in the
	// background: the message is sent again until it comes back, or the deadline is reached.
	deadline := time.After(30 * time.Second)
	retry := time.NewTicker(500 * time.Millisecond)
	defer retry.Stop()

	for {
		select {
		case got := <-received:
			if got != message {
				t.Fatalf("received %q, want %q", got, message)
			}

			return
		case <-retry.C:
			client.SendMessageToTopic(message, topic)
		case <-deadline:
			t.Fatal("the message was not consumed within 30s: is Kafka reachable, and auto-creation of topics enabled?")
		}
	}
}

// envOr returns the value of the environment variable name, or fallback if it is unset.
func envOr(t *testing.T, name, fallback string) string {
	t.Helper()

	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

// envIntOr returns the value of the environment variable name as an integer, or fallback if it is unset or invalid.
func envIntOr(t *testing.T, name string, fallback int) int {
	t.Helper()

	if value := os.Getenv(name); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}

	return fallback
}
