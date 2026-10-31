package ginamite

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/kafka/dlq"
	"github.com/cygy/ginamite/common/log"

	"github.com/IBM/sarama"
)

func TestKafkaOptions(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		options, err := kafkaOptions(config.NewConfiguration())

		if err != nil || options.TLS != nil || options.SASL != (kafka.SASL{}) || options.StatsInterval != 0 || options.Version != (sarama.KafkaVersion{}) {
			t.Errorf("options = %+v, error = %v", options, err)
		}
	})

	t.Run("all the settings", func(t *testing.T) {
		c := config.NewConfiguration()
		c.Kafka.Version = "3.6.0"
		c.Kafka.TLS.Enabled = true
		c.Kafka.SASL = config.KafkaSASLConfiguration{Mechanism: "scram-sha-512", Username: "user", Password: "secret"}
		c.Kafka.StatsInterval = 30

		options, err := kafkaOptions(c)
		if err != nil {
			t.Fatal(err)
		}
		if options.Version != sarama.V3_6_0_0 || options.TLS == nil || options.StatsInterval != 30*time.Second {
			t.Errorf("options = %+v", options)
		}
		if want := (kafka.SASL{Mechanism: kafka.SASLSCRAMSHA512, Username: "user", Password: "secret"}); options.SASL != want {
			t.Errorf("SASL = %+v, want %+v", options.SASL, want)
		}
	})

	t.Run("invalid settings", func(t *testing.T) {
		tests := map[string]func(c *config.Configuration){
			"version": func(c *config.Configuration) { c.Kafka.Version = "three" },
			"TLS": func(c *config.Configuration) {
				c.Kafka.TLS = config.KafkaTLSConfiguration{Enabled: true, CAFile: "missing.pem"}
			},
			"SASL mechanism": func(c *config.Configuration) {
				c.Kafka.SASL = config.KafkaSASLConfiguration{Mechanism: "KERBEROS", Username: "u"}
			},
			"SASL without. user": func(c *config.Configuration) { c.Kafka.SASL = config.KafkaSASLConfiguration{Mechanism: "PLAIN"} },
		}
		for name, change := range tests {
			c := config.NewConfiguration()
			change(c)

			if _, err := kafkaOptions(c); err == nil {
				t.Errorf("%s: error = nil", name)
			}
		}
	})
}

func TestCreateKafkaTopics(t *testing.T) {
	// Without Create, nothing is done: there is no broker to connect to.
	createKafkaTopics(config.NewConfiguration(), "127.0.0.1", 1, kafka.Options{})

	// A failure is logged, not fatal.
	c := config.NewConfiguration()
	c.Kafka.Topics.Create = true
	createKafkaTopics(c, "127.0.0.1", 1, kafka.Options{})
}

// fakeDLQ is a backend with two dead letters in a partition.
type fakeDLQ struct {
	next      int64
	published []string
	closed    bool
}

func (b *fakeDLQ) Partitions(string) ([]int32, error) { return []int32{0}, nil }

func (b *fakeDLQ) Position(string, int32) (int64, int64, error) { return b.next, 2, nil }

func (b *fakeDLQ) Read(_ context.Context, topic string, partition int32, from, to int64, fn func(*sarama.ConsumerMessage) error) error {
	for offset := from; offset < to; offset++ {
		message := &sarama.ConsumerMessage{
			Topic: topic, Partition: partition, Offset: offset, Timestamp: time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC),
			Value: []byte("secret body"),
			Headers: []*sarama.RecordHeader{
				{Key: []byte("error"), Value: []byte("smtp down")},
				{Key: []byte("topic"), Value: []byte("app_mail")},
				{Key: []byte("attempts"), Value: []byte("3")},
				{Key: []byte("group"), Value: []byte("app_mail")},
			},
		}
		if err := fn(message); err != nil {
			return err
		}
	}

	return nil
}

func (b *fakeDLQ) Commit(_ string, _ int32, next int64) error {
	b.next = next

	return nil
}

func (b *fakeDLQ) Publish(_ context.Context, message *sarama.ProducerMessage) error {
	b.published = append(b.published, message.Topic)

	return nil
}

func (b *fakeDLQ) Close() error {
	b.closed = true

	return nil
}

func TestServer_runDLQ(t *testing.T) {
	oldLog := log.Main
	t.Cleanup(func() { log.Main = oldLog })

	conf := writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "kafka_topics_prefix: app_\nservices:\n  kafka: {host: kafka, port: 9093}\nkafka: {stats_interval: 5}\n")

	var backend *fakeDLQ
	var gotHost string
	var gotPort int
	var gotGroup string
	old := openDLQ
	openDLQ = func(host string, port int, options kafka.Options, group string) (dlq.Backend, error) {
		gotHost, gotPort, gotGroup = host, port, group
		backend = &fakeDLQ{}
		if options.StatsInterval != 5*time.Second {
			t.Errorf("options = %+v", options)
		}

		return backend, nil
	}
	t.Cleanup(func() { openDLQ = old })

	run := func(command string, limit int, body bool) (string, error) {
		var out bytes.Buffer
		err := NewServer().runDLQ(&flags.Flags{ServerName: "dlq", Environment: "test", Version: "final", ServerConfigurationFile: conf, DLQ: command, DLQTopic: "mail", DLQLimit: limit, DLQBody: body}, &out)

		return out.String(), err
	}

	t.Run("list", func(t *testing.T) {
		out, err := run(flags.DLQList, 0, false)

		if err != nil || !backend.closed || gotHost != "kafka" || gotPort != 9093 || gotGroup != "app_mail_dlq_tool" {
			t.Fatalf("error = %v, backend = %+v, address = %s:%d, group = %q", err, backend, gotHost, gotPort, gotGroup)
		}
		for _, want := range []string{"app_mail_dlq[0]@0", "app_mail_dlq[0]@1", "2026-10-31T10:00:00Z", "from=app_mail[]@", "attempts=3", `error="smtp down"`, "size=11", "2 dead letters in app_mail_dlq"} {
			if !strings.Contains(out, want) {
				t.Errorf("the output does not contain %q:\n%s", want, out)
			}
		}
		if strings.Contains(out, "secret body") {
			t.Errorf("the body must not be printed:\n%s", out)
		}
		if backend.next != 0 {
			t.Errorf("listing moved the position to %d", backend.next)
		}
	})

	t.Run("list with the body", func(t *testing.T) {
		if out, err := run(flags.DLQList, 1, true); err != nil || !strings.Contains(out, "    secret body") || !strings.Contains(out, "1 dead letters in") {
			t.Errorf("error = %v, output:\n%s", err, out)
		}
	})

	t.Run("replay", func(t *testing.T) {
		out, err := run(flags.DLQReplay, 0, false)

		if err != nil || len(backend.published) != 2 || backend.published[0] != "app_mail" || backend.next != 2 {
			t.Errorf("error = %v, published = %v, position = %d", err, backend.published, backend.next)
		}
		if !strings.Contains(out, "app_mail_dlq[0]@0 sent again to app_mail") || !strings.Contains(out, "2 dead letters sent again from app_mail_dlq") {
			t.Errorf("output:\n%s", out)
		}
	})

	t.Run("purge", func(t *testing.T) {
		out, err := run(flags.DLQPurge, 0, false)

		if err != nil || backend.next != 2 || len(backend.published) != 0 || !strings.Contains(out, "2 dead letters purged from app_mail_dlq") {
			t.Errorf("error = %v, position = %d, output:\n%s", err, backend.next, out)
		}
	})

	t.Run("failures", func(t *testing.T) {
		openDLQ = func(string, int, kafka.Options, string) (dlq.Backend, error) { return nil, errors.New("no broker") }
		if _, err := run(flags.DLQList, 0, false); err == nil {
			t.Error("a connection failure must be returned")
		}

		noKafka := writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "kafka_topics_prefix: app_\n")
		if err := NewServer().runDLQ(&flags.Flags{ServerName: "dlq", Environment: "test", Version: "final", ServerConfigurationFile: noKafka, DLQ: flags.DLQList, DLQTopic: "mail"}, &bytes.Buffer{}); err == nil {
			t.Error("a configuration without kafka service must be refused")
		}

		badOptions := writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "services:\n  kafka: {host: kafka, port: 1}\nkafka: {version: three}\n")
		if err := NewServer().runDLQ(&flags.Flags{ServerName: "dlq", Environment: "test", Version: "final", ServerConfigurationFile: badOptions, DLQ: flags.DLQList, DLQTopic: "mail"}, &bytes.Buffer{}); err == nil {
			t.Error("invalid options must be refused")
		}
	})
}
