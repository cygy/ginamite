package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/flags"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/kafka/dlq"
)

// openDLQ connects to the topics of the dead letters; the tests replace it.
var openDLQ = dlq.Open

// Command lists, replays or purges the dead letters of the topic of the arguments, and writes the report to out, if the
// arguments ask for it, see flags.Flags.DLQ. No server is started: it only needs the configuration of the server and of
// Kafka. It makes the plugin a ginamite.Commander.
func (p *Plugin) Command(f *flags.Flags, out io.Writer) (handled bool, err error) {
	if f.DLQ == "" {
		return false, nil
	}

	return true, runDLQ(f, out)
}

// runDLQ runs the command on the dead letters of the arguments.
func runDLQ(f *flags.Flags, out io.Writer) error {
	log.Initialize(f.ServerName, f.Environment == environment.Production)

	configuration := config.NewConfiguration()
	if err := configuration.Initialize(f.ServerConfigurationFile, f.Environment, f.Version); err != nil {
		return err
	}

	kafkaConf, ok := configuration.KafkaConfiguration()
	if !ok {
		return errors.New("the configuration has no kafka service")
	}

	options, err := optionsFrom(configuration)
	if err != nil {
		return err
	}

	topic := DeadLetterTopic(configuration.KafkaTopicsPrefix + f.DLQTopic)

	backend, err := openDLQ(kafkaConf.Host, kafkaConf.Port, options.NewConfig(), topic+"_tool")
	if err != nil {
		return err
	}
	defer func() { _ = backend.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch f.DLQ {
	case flags.DLQList:
		count, err := dlq.List(ctx, backend, topic, f.DLQLimit, func(e dlq.Entry) {
			_, _ = fmt.Fprintf(out, "%s  %s  from=%s[%s]@%s  group=%s  attempts=%s  size=%d  error=%q\n",
				e.ID(), e.Time.UTC().Format("2006-01-02T15:04:05Z"), e.OriginalTopic, e.OriginalPartition, e.OriginalOffset, e.Group, e.Attempts, len(e.Value), e.Error)
			if f.DLQBody {
				_, _ = fmt.Fprintf(out, "    %s\n", e.Value)
			}
		})
		_, _ = fmt.Fprintf(out, "%d dead letters in %s\n", count, topic)

		return err

	case flags.DLQReplay:
		count, err := dlq.Replay(ctx, backend, topic, f.DLQLimit, func(e dlq.Entry) {
			_, _ = fmt.Fprintf(out, "%s sent again to %s\n", e.ID(), e.OriginalTopic)
		})
		_, _ = fmt.Fprintf(out, "%d dead letters sent again from %s\n", count, topic)

		return err

	default:
		count, err := dlq.Purge(ctx, backend, topic, f.DLQLimit)
		_, _ = fmt.Fprintf(out, "%d dead letters purged from %s\n", count, topic)

		return err
	}
}
