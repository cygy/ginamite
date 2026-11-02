package dlq

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

const testTopic = "app_mail_dlq"

// newBroker returns a broker with the topic of the dead letters, which has one partition holding three messages, and
// the group of the tool which saved the position 1.
func newBroker(t *testing.T, savedPosition int64) *sarama.MockBroker {
	t.Helper()

	broker := sarama.NewMockBroker(t, 1)
	t.Cleanup(broker.Close)

	fetch := sarama.NewMockFetchResponse(t, 1)
	for offset, body := range []string{"zero", "one", "two"} {
		fetch.SetMessage(testTopic, 0, int64(offset), sarama.StringEncoder(body))
	}

	broker.SetHandlerByMap(map[string]sarama.MockResponse{
		"ApiVersionsRequest": sarama.NewMockApiVersionsResponse(t).SetApiKeys([]sarama.ApiVersionsResponseKey{
			{ApiKey: 0, MinVersion: 0, MaxVersion: 8},  // Produce
			{ApiKey: 1, MinVersion: 0, MaxVersion: 11}, // Fetch
			{ApiKey: 2, MinVersion: 0, MaxVersion: 5},  // ListOffsets
			{ApiKey: 3, MinVersion: 0, MaxVersion: 8},  // Metadata
			{ApiKey: 8, MinVersion: 0, MaxVersion: 7},  // OffsetCommit
			{ApiKey: 9, MinVersion: 0, MaxVersion: 5},  // OffsetFetch
			{ApiKey: 10, MinVersion: 0, MaxVersion: 3}, // FindCoordinator
			{ApiKey: 22, MinVersion: 0, MaxVersion: 3}, // InitProducerID
		}),
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(broker.Addr(), broker.BrokerID()).
			SetController(broker.BrokerID()).
			SetLeader(testTopic, 0, broker.BrokerID()).
			SetLeader("app_mail", 0, broker.BrokerID()),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset(testTopic, 0, sarama.OffsetOldest, 0).
			SetOffset(testTopic, 0, sarama.OffsetNewest, 3),
		"FetchRequest":           fetch,
		"FindCoordinatorRequest": sarama.NewMockFindCoordinatorResponse(t).SetCoordinator(sarama.CoordinatorGroup, "tool", broker),
		"OffsetFetchRequest":     sarama.NewMockOffsetFetchResponse(t).SetOffset("tool", testTopic, 0, savedPosition, "", sarama.ErrNoError),
		"OffsetCommitRequest":    sarama.NewMockOffsetCommitResponse(t),
		"InitProducerIDRequest":  sarama.NewMockInitProducerIDResponse(t),
		"ProduceRequest":         sarama.NewMockProduceResponse(t).SetError("app_mail", 0, sarama.ErrNoError),
	})

	return broker
}

func openBackend(t *testing.T, broker *sarama.MockBroker) Backend {
	t.Helper()

	host, port, _ := strings.Cut(broker.Addr(), ":")
	var portNumber int
	for _, c := range port {
		portNumber = portNumber*10 + int(c-'0')
	}

	b, err := Open(host, portNumber, saramaConfig(sarama.V2_8_0_0), "tool")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = b.Close() })

	return b
}

func TestOpen_Failure(t *testing.T) {
	if _, err := Open("127.0.0.1", 1, saramaConfig(sarama.V2_8_0_0), "tool"); err == nil {
		t.Error("Open() must fail without broker")
	}
}

func TestBackend(t *testing.T) {
	ctx := context.Background()

	t.Run("positions", func(t *testing.T) {
		b := openBackend(t, newBroker(t, 1))

		partitions, err := b.Partitions(testTopic)
		if err != nil || len(partitions) != 1 || partitions[0] != 0 {
			t.Fatalf("partitions = %v, error = %v", partitions, err)
		}

		next, end, err := b.Position(testTopic, 0)
		if err != nil || next != 1 || end != 3 {
			t.Errorf("position = %d..%d, error = %v, want the saved position 1 and the end 3", next, end, err)
		}
	})

	t.Run("no saved position starts at the oldest message", func(t *testing.T) {
		b := openBackend(t, newBroker(t, -1))

		next, _, err := b.Position(testTopic, 0)
		if err != nil || next != 0 {
			t.Errorf("next = %d, error = %v", next, err)
		}
	})

	t.Run("reads the messages in the range", func(t *testing.T) {
		b := openBackend(t, newBroker(t, 0))

		var bodies []string
		err := b.Read(ctx, testTopic, 0, 1, 3, func(m *sarama.ConsumerMessage) error {
			bodies = append(bodies, string(m.Value))

			return nil
		})

		if err != nil || strings.Join(bodies, ",") != "one,two" {
			t.Errorf("bodies = %v, error = %v", bodies, err)
		}
	})

	t.Run("stops at the error of the function", func(t *testing.T) {
		b := openBackend(t, newBroker(t, 0))

		count := 0
		err := b.Read(ctx, testTopic, 0, 0, 3, func(*sarama.ConsumerMessage) error {
			count++

			return context.DeadlineExceeded
		})

		if !errors.Is(err, context.DeadlineExceeded) || count != 1 {
			t.Errorf("count = %d, error = %v", count, err)
		}
	})

	t.Run("an interrupted read", func(t *testing.T) {
		b := openBackend(t, newBroker(t, 0))

		canceled, cancel := context.WithCancel(ctx)
		cancel()

		err := b.Read(canceled, testTopic, 0, 0, 3, func(*sarama.ConsumerMessage) error { return nil })
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("saves the position", func(t *testing.T) {
		b := openBackend(t, newBroker(t, 0))

		if err := b.Commit(testTopic, 0, 2); err != nil {
			t.Errorf("Commit() error = %v", err)
		}
	})

	t.Run("publishes", func(t *testing.T) {
		b := openBackend(t, newBroker(t, 0))

		err := b.Publish(ctx, &sarama.ProducerMessage{Topic: "app_mail", Value: sarama.StringEncoder("body")})
		if err != nil {
			t.Errorf("Publish() error = %v", err)
		}

		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if err := b.Publish(canceled, &sarama.ProducerMessage{Topic: "app_mail", Value: sarama.StringEncoder("body")}); err == nil {
			t.Error("Publish() must fail when interrupted")
		}
	})

	t.Run("unknown topic", func(t *testing.T) {
		b := openBackend(t, newBroker(t, 0))

		if _, err := b.Partitions("unknown"); err == nil {
			t.Error("Partitions() must fail for an unknown topic")
		}
	})
}

func TestBackend_Timeout(t *testing.T) {
	// A message which does not come is waited for a limited time.
	old := idleTimeout
	idleTimeout = 50 * time.Millisecond
	t.Cleanup(func() { idleTimeout = old })

	b := openBackend(t, newBroker(t, 0))

	// The offset 7 does not exist in the topic.
	err := b.Read(context.Background(), testTopic, 0, 7, 9, func(*sarama.ConsumerMessage) error { return nil })
	if err == nil {
		t.Error("Read() must fail when no message comes")
	}
}

// saramaConfig returns a configuration for the brokers of the version.
func saramaConfig(version sarama.KafkaVersion) *sarama.Config {
	config := sarama.NewConfig()
	config.Version = version
	config.Producer.Return.Successes = true

	return config
}
