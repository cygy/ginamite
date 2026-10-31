package kafka

import (
	"maps"
	"slices"
	"strings"

	"github.com/cygy/ginamite/common/metrics"
)

// Collect writes the counters of the client, the number of messages waiting to be sent and the lag of each partition,
// see Stats. It makes the client a metrics.Collector.
func (client *Client) Collect(w *metrics.Writer) {
	stats := client.Stats()

	w.Counter("kafka_messages_received_total", "Messages consumed.", float64(stats.Received))
	w.Counter("kafka_messages_handled_total", "Messages handled successfully.", float64(stats.Handled))
	w.Counter("kafka_messages_retried_total", "Attempts to handle a message after a failure.", float64(stats.Retried))
	w.Counter("kafka_messages_dead_lettered_total", "Messages which could not be handled and went to their dead letter topic.", float64(stats.DeadLettered))
	w.Counter("kafka_messages_duplicates_total", "Messages ignored because they were already handled.", float64(stats.Duplicates))
	w.Counter("kafka_messages_sent_total", "Messages acknowledged by Kafka.", float64(stats.Sent))
	w.Counter("kafka_send_errors_total", "Messages refused by Kafka.", float64(stats.SendErrors))
	w.Counter("kafka_messages_dropped_total", "Messages lost because the client is closed or its queue is full.", float64(stats.Dropped))
	w.Gauge("kafka_outbox_messages", "Messages waiting to be sent.", float64(len(client.outbox)))

	for _, partition := range slices.Sorted(maps.Keys(stats.Lag)) {
		topic, number, _ := strings.Cut(partition, "/")
		w.Gauge("kafka_consumer_lag_messages", "Messages waiting to be consumed, as seen when the last message of the partition was handled.",
			float64(stats.Lag[partition]), "topic", topic, "partition", number)
	}
}
