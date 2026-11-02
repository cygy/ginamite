package kafka

import (
	"context"
	"errors"
	"flag"
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

// The commands on the dead letters.
const (
	DLQList   = "list"
	DLQReplay = "replay"
	DLQPurge  = "purge"
)

// dlqArguments are the arguments of the command on the dead letters.
type dlqArguments struct {
	command string // list, replay or purge. No server is started if set.
	topic   string // the topic, without the prefix, whose dead letters the command works on.
	limit   int    // the maximum count of dead letters the command works on, all of them if 0.
	body    bool   // whether the list prints the body of the messages, which holds secrets.
}

// RegisterFlags adds the flags of the command on the dead letters: -dlq, -dlq-topic, -dlq-limit and -dlq-body. It
// makes the plugin a ginamite.Commander.
func (p *Plugin) RegisterFlags(set *flag.FlagSet) {
	set.StringVar(&p.dlq.command, "dlq", "", fmt.Sprintf("Command on the dead letters of a topic (%s, %s or %s), instead of starting a server. Needs -name, -conf-server and -dlq-topic.", DLQList, DLQReplay, DLQPurge))
	set.StringVar(&p.dlq.topic, "dlq-topic", "", "Topic, without the prefix, whose dead letters the -dlq command works on.")
	set.IntVar(&p.dlq.limit, "dlq-limit", 0, "Maximum count of dead letters the -dlq command works on (all if 0).")
	set.BoolVar(&p.dlq.body, "dlq-body", false, "Print the body of the dead letters with -dlq list. It holds secrets.")
}

// Command lists, replays or purges the dead letters of the topic of the arguments, and writes the report to out, if the
// arguments ask for it with -dlq. No server is started: it only needs the configuration of the server and of Kafka. It
// returns an error if the arguments are invalid. It makes the plugin a ginamite.Commander.
func (p *Plugin) Command(f *flags.Flags, out io.Writer) (handled bool, err error) {
	if p.dlq.command == "" {
		return false, nil
	}

	if err := p.dlq.validate(f); err != nil {
		return true, err
	}

	return true, runDLQ(f, p.dlq, out)
}

// validate checks the arguments of a command on the dead letters.
func (a dlqArguments) validate(f *flags.Flags) error {
	switch {
	case a.command != DLQList && a.command != DLQReplay && a.command != DLQPurge:
		return fmt.Errorf("the command -dlq must be '%s', '%s' or '%s': show the help to learn more about", DLQList, DLQReplay, DLQPurge)
	case f.IsAPIServer || f.IsWorkerServer || f.IsWebServer:
		return errors.New("the command -dlq starts no server: do not use -api, -worker nor -web")
	case len(a.topic) == 0:
		return errors.New("a topic must be defined with -dlq-topic: show the help to learn more about")
	case a.limit < 0:
		return errors.New("the limit -dlq-limit must not be negative")
	case len(f.ServerName) == 0:
		return errors.New("a server name must be defined: show the help to learn more about")
	case len(f.ServerConfigurationFile) == 0:
		return errors.New("a path to the server configuration file must be defined: show the help to learn more about")
	}

	return nil
}

// runDLQ runs the command on the dead letters of the arguments.
func runDLQ(f *flags.Flags, a dlqArguments, out io.Writer) error {
	log.Initialize(f.ServerName, f.Environment == environment.Production)

	configuration := config.NewConfiguration()
	if err := configuration.Initialize(f.ServerConfigurationFile, f.Environment, f.Version); err != nil {
		return err
	}

	if !enabled(configuration) {
		return errors.New("the configuration has no kafka broker")
	}

	settings, err := settingsFrom(configuration)
	if err != nil {
		return err
	}

	options, err := optionsFrom(settings)
	if err != nil {
		return err
	}

	topic := DeadLetterTopic(settings.TopicsPrefix + a.topic)

	backend, err := openDLQ(configuration.Broker.Host, configuration.Broker.Port, options.NewConfig(), topic+"_tool")
	if err != nil {
		return err
	}
	defer func() { _ = backend.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch a.command {
	case DLQList:
		count, err := dlq.List(ctx, backend, topic, a.limit, func(e dlq.Entry) {
			_, _ = fmt.Fprintf(out, "%s  %s  from=%s[%s]@%s  group=%s  attempts=%s  size=%d  error=%q\n",
				e.ID(), e.Time.UTC().Format("2006-01-02T15:04:05Z"), e.OriginalTopic, e.OriginalPartition, e.OriginalOffset, e.Group, e.Attempts, len(e.Value), e.Error)
			if a.body {
				_, _ = fmt.Fprintf(out, "    %s\n", e.Value)
			}
		})
		_, _ = fmt.Fprintf(out, "%d dead letters in %s\n", count, topic)

		return err

	case DLQReplay:
		count, err := dlq.Replay(ctx, backend, topic, a.limit, func(e dlq.Entry) {
			_, _ = fmt.Fprintf(out, "%s sent again to %s\n", e.ID(), e.OriginalTopic)
		})
		_, _ = fmt.Fprintf(out, "%d dead letters sent again from %s\n", count, topic)

		return err

	default:
		count, err := dlq.Purge(ctx, backend, topic, a.limit)
		_, _ = fmt.Fprintf(out, "%d dead letters purged from %s\n", count, topic)

		return err
	}
}
