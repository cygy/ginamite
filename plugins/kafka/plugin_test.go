package kafka

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/kafka/internal/dlq"

	"github.com/IBM/sarama"
)

func TestOptionsFrom(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		options, err := optionsFrom(DefaultSettings())

		if err != nil || options.TLS != nil || options.SASL != (SASL{}) || options.StatsInterval != 0 || options.Version != (sarama.KafkaVersion{}) {
			t.Errorf("options = %+v, error = %v", options, err)
		}
	})

	t.Run("all the settings", func(t *testing.T) {
		k := DefaultSettings()
		k.Version = "3.6.0"
		k.TLS.Enabled = true
		k.SASL = SASLSettings{Mechanism: "scram-sha-512", Username: "user", Password: "secret"}
		k.StatsInterval = 30

		options, err := optionsFrom(k)
		if err != nil {
			t.Fatal(err)
		}
		if options.Version != sarama.V3_6_0_0 || options.TLS == nil || options.StatsInterval != 30*time.Second {
			t.Errorf("options = %+v", options)
		}
		if want := (SASL{Mechanism: SASLSCRAMSHA512, Username: "user", Password: "secret"}); options.SASL != want {
			t.Errorf("SASL = %+v, want %+v", options.SASL, want)
		}
	})

	t.Run("invalid settings", func(t *testing.T) {
		tests := map[string]func(k *Settings){
			"version": func(k *Settings) { k.Version = "three" },
			"TLS": func(k *Settings) {
				k.TLS = TLSSettings{Enabled: true, CAFile: "missing.pem"}
			},
			"SASL mechanism": func(k *Settings) {
				k.SASL = SASLSettings{Mechanism: "KERBEROS", Username: "u"}
			},
			"SASL without. user": func(k *Settings) { k.SASL = SASLSettings{Mechanism: "PLAIN"} },
		}
		for name, change := range tests {
			k := DefaultSettings()
			change(&k)

			if _, err := optionsFrom(k); err == nil {
				t.Errorf("%s: error = nil", name)
			}
		}
	})
}

func TestCreateTopics(t *testing.T) {
	// Without Create, nothing is done: there is no broker to connect to.
	createTopics(DefaultSettings(), "127.0.0.1", 1, Options{})

	// A failure is logged, not fatal.
	k := DefaultSettings()
	k.Topics.Create = true
	createTopics(k, "127.0.0.1", 1, Options{})
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

func TestRunDLQ(t *testing.T) {
	oldLog := log.Main
	t.Cleanup(func() { log.Main = oldLog })

	conf := writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "broker: {service: kafka, host: kafka, port: 9093}\nkafka: {topics_prefix: app_, stats_interval: 5}\n")

	var backend *fakeDLQ
	var gotHost string
	var gotPort int
	var gotGroup string
	old := openDLQ
	openDLQ = func(host string, port int, config *sarama.Config, group string) (dlq.Backend, error) {
		gotHost, gotPort, gotGroup = host, port, group
		backend = &fakeDLQ{}
		if config == nil {
			t.Error("no configuration")
		}

		return backend, nil
	}
	t.Cleanup(func() { openDLQ = old })

	run := func(command string, limit int, body bool) (string, error) {
		var out bytes.Buffer
		err := runDLQ(t.Context(), &flags.Flags{ServerName: "dlq", Environment: "test", Version: "final", ServerConfigurationFile: conf}, dlqArguments{command: command, topic: "mail", limit: limit, body: body}, &out)

		return out.String(), err
	}

	t.Run("list", func(t *testing.T) {
		out, err := run(DLQList, 0, false)

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
		if out, err := run(DLQList, 1, true); err != nil || !strings.Contains(out, "    secret body") || !strings.Contains(out, "1 dead letters in") {
			t.Errorf("error = %v, output:\n%s", err, out)
		}
	})

	t.Run("replay", func(t *testing.T) {
		out, err := run(DLQReplay, 0, false)

		if err != nil || len(backend.published) != 2 || backend.published[0] != "app_mail" || backend.next != 2 {
			t.Errorf("error = %v, published = %v, position = %d", err, backend.published, backend.next)
		}
		if !strings.Contains(out, "app_mail_dlq[0]@0 sent again to app_mail") || !strings.Contains(out, "2 dead letters sent again from app_mail_dlq") {
			t.Errorf("output:\n%s", out)
		}
	})

	t.Run("purge", func(t *testing.T) {
		out, err := run(DLQPurge, 0, false)

		if err != nil || backend.next != 2 || len(backend.published) != 0 || !strings.Contains(out, "2 dead letters purged from app_mail_dlq") {
			t.Errorf("error = %v, position = %d, output:\n%s", err, backend.next, out)
		}
	})

	t.Run("failures", func(t *testing.T) {
		openDLQ = func(string, int, *sarama.Config, string) (dlq.Backend, error) { return nil, errors.New("no broker") }
		if _, err := run(DLQList, 0, false); err == nil {
			t.Error("a connection failure must be returned")
		}

		noKafka := writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "broker: {service: nats, host: nats, port: 4222}\nkafka: {topics_prefix: app_}\n")
		if err := runDLQ(t.Context(), &flags.Flags{ServerName: "dlq", Environment: "test", Version: "final", ServerConfigurationFile: noKafka}, dlqArguments{command: DLQList, topic: "mail"}, &bytes.Buffer{}); err == nil {
			t.Error("a configuration without kafka service must be refused")
		}

		badOptions := writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "broker: {service: kafka, host: kafka, port: 1}\nkafka: {version: three}\n")
		if err := runDLQ(t.Context(), &flags.Flags{ServerName: "dlq", Environment: "test", Version: "final", ServerConfigurationFile: badOptions}, dlqArguments{command: DLQList, topic: "mail"}, &bytes.Buffer{}); err == nil {
			t.Error("invalid options must be refused")
		}
	})
}

func writeFile(t *testing.T, path, content string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestPlugin_Command(t *testing.T) {
	parse := func(t *testing.T, args ...string) (*Plugin, *flags.Flags, error) {
		t.Helper()

		p := NewPlugin()
		f := flags.NewFlags()

		return p, f, f.Parse(args, &bytes.Buffer{}, p.RegisterFlags)
	}

	t.Run("nothing to do without a command", func(t *testing.T) {
		p, f, err := parse(t, "-api", "-name", "n")
		if err != nil {
			t.Fatal(err)
		}

		if handled, err := p.Command(t.Context(), f, &bytes.Buffer{}); handled || err != nil {
			t.Errorf("Command() = %v, %v", handled, err)
		}
	})

	t.Run("the flags", func(t *testing.T) {
		p, _, err := parse(t, "-dlq", "replay", "-dlq-topic", "mail", "-dlq-limit", "5", "-dlq-body")

		if want := (dlqArguments{command: DLQReplay, topic: "mail", limit: 5, body: true}); err != nil || p.dlq != want {
			t.Errorf("Parse() = %v, arguments = %+v, want %+v", err, p.dlq, want)
		}
	})

	t.Run("failure of the command", func(t *testing.T) {
		noKafka := writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "kafka: {topics_prefix: app_}\n")
		p, f, err := parse(t, "-dlq", "list", "-dlq-topic", "mail", "-name", "dlq", "-conf-server", noKafka)
		if err != nil {
			t.Fatal(err)
		}

		handled, err := p.Command(t.Context(), f, &bytes.Buffer{})
		if !handled || err == nil {
			t.Errorf("Command() = %v, %v, want the command handled and its failure", handled, err)
		}
	})

	invalid := map[string]struct {
		args []string
		want string
	}{
		"unknown command":       {[]string{"-dlq", "delete", "-dlq-topic", "mail", "-name", "n", "-conf-server", "s"}, "must be 'list', 'replay' or 'purge'"},
		"with a server":         {[]string{"-dlq", "list", "-api", "-dlq-topic", "mail", "-name", "n", "-conf-server", "s"}, "starts no server"},
		"without topic":         {[]string{"-dlq", "list", "-name", "n", "-conf-server", "s"}, "-dlq-topic"},
		"with a negative limit": {[]string{"-dlq", "list", "-dlq-topic", "mail", "-dlq-limit", "-1", "-name", "n", "-conf-server", "s"}, "must not be negative"},
		"without name":          {[]string{"-dlq", "list", "-dlq-topic", "mail", "-conf-server", "s"}, "server name"},
		"without configuration": {[]string{"-dlq", "list", "-dlq-topic", "mail", "-name", "n"}, "server configuration file"},
	}
	for name, tt := range invalid {
		t.Run(name, func(t *testing.T) {
			p, f, err := parse(t, tt.args...)
			if err != nil {
				t.Fatal(err)
			}

			handled, err := p.Command(t.Context(), f, &bytes.Buffer{})
			if !handled || err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Command() = %v, %v, want the command handled with an error about %q", handled, err, tt.want)
			}
		})
	}
}

func TestEnabled(t *testing.T) {
	tests := []struct {
		name   string
		broker config.BrokerConfiguration
		want   bool
	}{
		{"nothing", config.BrokerConfiguration{}, false},
		{"another broker", config.BrokerConfiguration{Service: "nats", Host: "nats", Port: 4222}, false},
		{"no host", config.BrokerConfiguration{Service: "kafka", Port: 9092}, false},
		{"no port", config.BrokerConfiguration{Service: "kafka", Host: "kafka"}, false},
		{"configured", config.BrokerConfiguration{Service: "kafka", Host: "kafka", Port: 9092}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := config.NewConfiguration()
			c.Broker = tt.broker

			if got := enabled(c); got != tt.want {
				t.Errorf("enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlugin_Configure(t *testing.T) {
	t.Run("without service", func(t *testing.T) {
		s := ginamite.NewServer()
		if err := NewPlugin().Configure(s); err != nil {
			t.Fatal(err)
		}

		if s.Broker != nil {
			t.Errorf("broker = %v, want none without the address of kafka", s.Broker)
		}
	})

	t.Run("with service", func(t *testing.T) {
		s := ginamite.NewServer()
		if err := s.Config.Initialize(writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "broker: {service: kafka, host: kafka, port: 9093}\nkafka: {topics_prefix: app_, max_attempts: 4, retry_delay: 250}\n"), "test", "final"); err != nil {
			t.Fatal(err)
		}
		if err := NewPlugin().Configure(s); err != nil {
			t.Fatal(err)
		}

		client, ok := s.Broker.(*Client)
		if !ok {
			t.Fatalf("broker = %T, want the Kafka client", s.Broker)
		}
		if client.Host != "kafka" || client.Port != 9093 || client.TopicsPrefix != "app_" || client.MaxAttempts != 4 || client.RetryDelay != 250*time.Millisecond {
			t.Errorf("client = %+v", client)
		}
		if s.Queue == nil || s.API.Queue != s.Queue {
			t.Error("the queue must send through the client")
		}

		var _ broker.DeduplicatorUser = client
	})

	t.Run("invalid settings", func(t *testing.T) {
		s := ginamite.NewServer()
		if err := s.Config.Initialize(writeFile(t, filepath.Join(t.TempDir(), "server.yaml"), "broker: {service: kafka, host: kafka, port: 9093}\nkafka: {version: not a version}\n"), "test", "final"); err != nil {
			t.Fatal(err)
		}

		if err := NewPlugin().Configure(s); err == nil || s.Broker != nil {
			t.Errorf("Configure() = %v, broker = %v, want an error and no broker", err, s.Broker)
		}
	})
}
