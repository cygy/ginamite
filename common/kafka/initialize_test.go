package kafka

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

// newMockBroker returns a broker answering the metadata, offset and fetch requests of a single-partition topic.
func newMockBroker(t *testing.T, topic string) *sarama.MockBroker {
	t.Helper()

	broker := sarama.NewMockBroker(t, 1)
	t.Cleanup(broker.Close)

	broker.SetHandlerByMap(map[string]sarama.MockResponse{
		"ApiVersionsRequest": sarama.NewMockApiVersionsResponse(t).SetApiKeys([]sarama.ApiVersionsResponseKey{
			{ApiKey: 0, MinVersion: 0, MaxVersion: 8},  // Produce
			{ApiKey: 1, MinVersion: 0, MaxVersion: 11}, // Fetch
			{ApiKey: 2, MinVersion: 0, MaxVersion: 5},  // ListOffsets
			{ApiKey: 3, MinVersion: 0, MaxVersion: 8},  // Metadata
		}),
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(broker.Addr(), broker.BrokerID()).
			SetLeader(topic, 0, broker.BrokerID()),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset(topic, 0, sarama.OffsetNewest, 0).
			SetOffset(topic, 0, sarama.OffsetOldest, 0),
		"FetchRequest": sarama.NewMockFetchResponse(t, 1).
			SetMessage(topic, 0, 0, sarama.StringEncoder("consumed")),
	})

	return broker
}

func splitAddress(t *testing.T, addr string) (string, int) {
	t.Helper()

	host, port, found := strings.Cut(addr, ":")
	if !found {
		t.Fatalf("invalid address %q", addr)
	}

	n, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	return host, n
}

func TestClient_Initialize_Consumer(t *testing.T) {
	hook := useHook(t)
	broker := newMockBroker(t, "app_users")
	host, port := splitAddress(t, broker.Addr())

	received := make(chan []byte, 1)
	client := NewClient()
	client.Initialize(host, port, "app_", false, []MessageConsumer{{
		Topic:  "users",
		Action: func(value []byte) { received <- value },
	}})

	if client.HasProducer || client.TopicsPrefix != "app_" {
		t.Errorf("client = %+v", client)
	}

	select {
	case value := <-received:
		if string(value) != "consumed" {
			t.Errorf("value = %q", value)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the message was not consumed")
	}

	waitForLog(t, hook, "kafka message received")
}

func TestClient_Initialize_Producer(t *testing.T) {
	hook := useHook(t)
	broker := newMockBroker(t, "app_users")
	host, port := splitAddress(t, broker.Addr())

	client := NewClient()
	client.Initialize(host, port, "app_", true, nil)

	if !client.HasProducer {
		t.Fatal("the client must be a producer")
	}

	waitForLog(t, hook, "connected to kafka")
}
