// Package flags parses the command line arguments of the servers.
package flags

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"
)

// Flags holds the command line arguments passed at launch.
type Flags struct {
	IsAPIServer                  bool
	IsWorkerServer               bool
	IsWebServer                  bool
	ServerName                   string
	Environment                  string
	Version                      string
	ServerConfigurationFile      string
	ApplicationConfigurationFile string
	RoutesFile                   string
	Help                         bool

	// DLQ is the command on the dead letters of a topic: "list", "replay" or "purge". No server is started if set.
	DLQ string
	// DLQTopic is the topic, without the prefix, whose dead letters the command works on.
	DLQTopic string
	// DLQLimit is the maximum count of dead letters the command works on, all of them if 0.
	DLQLimit int
	// DLQBody tells the list to print the body of the messages, which holds secrets.
	DLQBody bool
}

// DLQ commands.
const (
	DLQList   = "list"
	DLQReplay = "replay"
	DLQPurge  = "purge"
)

// NewFlags returns an empty Flags.
func NewFlags() *Flags {
	return &Flags{}
}

// Parse parses the command line arguments of the process and reports whether a server or a command must run.
// It prints the usage and returns false if the help is requested.
// It returns an error if an argument is invalid or a mandatory one is missing; the application configuration file is
// mandatory only if applicationConfigurationRequired is true.
func (f *Flags) Parse(applicationConfigurationRequired bool) (bool, error) {
	ok, problem := f.parse(os.Args[1:], os.Stderr, applicationConfigurationRequired)
	if problem != "" {
		return false, errors.New(problem)
	}

	return ok, nil
}

// parse parses args and reports whether they are valid.
// The usage is written to output. When the arguments are invalid, the returned
// message describes the first problem.
func (f *Flags) parse(args []string, output io.Writer, applicationConfigurationRequired bool) (bool, string) {
	set := flag.NewFlagSet("server", flag.ContinueOnError)
	set.SetOutput(output)

	isAPIServer := set.Bool("api", false, "Present if the server is a API server.")
	isWorkerServer := set.Bool("worker", false, "Present if the server is a worker server.")
	isWebServer := set.Bool("web", false, "Present if the server is a web server.")
	serverName := set.String("name", "", "Name of the server.")
	env := set.String("env", environment.Development, fmt.Sprintf("Running environment of the server (%s, %s or %s)", environment.Development, environment.Production, environment.Test))
	ver := set.String("version", version.Final, fmt.Sprintf("Running version of the server (%s, %s or %s)", version.Final, version.Beta, version.Alpha))
	serverConfigurationFile := set.String("conf-server", "", "Path to the server configuration file.")
	applicationConfigurationFile := set.String("conf-app", "", "Path to the application configuration file.")
	routesFile := set.String("routes", "", "Path to the web routes file.")
	help := set.Bool("help", false, "Print the help message")
	dlq := set.String("dlq", "", fmt.Sprintf("Command on the dead letters of a topic (%s, %s or %s), instead of starting a server. Needs -name, -conf-server and -dlq-topic.", DLQList, DLQReplay, DLQPurge))
	dlqTopic := set.String("dlq-topic", "", "Topic, without the prefix, whose dead letters the -dlq command works on.")
	dlqLimit := set.Int("dlq-limit", 0, "Maximum count of dead letters the -dlq command works on (all if 0).")
	dlqBody := set.Bool("dlq-body", false, "Print the body of the dead letters with -dlq list. It holds secrets.")

	if err := set.Parse(args); err != nil {
		return false, err.Error()
	}

	f.IsAPIServer = *isAPIServer
	f.IsWorkerServer = *isWorkerServer
	f.IsWebServer = *isWebServer
	f.ServerName = *serverName
	f.Environment = *env
	f.Version = *ver
	f.ServerConfigurationFile = *serverConfigurationFile
	f.ApplicationConfigurationFile = *applicationConfigurationFile
	f.RoutesFile = *routesFile
	f.Help = *help
	f.DLQ = *dlq
	f.DLQTopic = *dlqTopic
	f.DLQLimit = *dlqLimit
	f.DLQBody = *dlqBody

	if f.Help {
		set.PrintDefaults()
		return false, ""
	}

	switch {
	case f.DLQ != "":
		return f.validateDLQ()
	case !f.IsAPIServer && !f.IsWorkerServer && !f.IsWebServer:
		return false, "The server must be defined as 'api', 'worker' or 'web'. Show the help to learn more about."
	case len(f.ServerName) == 0:
		return false, "A server name must be defined. Show the help to learn more about."
	case len(f.ServerConfigurationFile) == 0:
		return false, "A path to the server configuration file must be defined. Show the help to learn more about."
	case applicationConfigurationRequired && len(f.ApplicationConfigurationFile) == 0:
		return false, "A path to the application configuration file must be defined. Show the help to learn more about."
	case f.IsWebServer && len(f.RoutesFile) == 0:
		return false, "A path to the web routes file must be defined. Show the help to learn more about."
	}

	return true, ""
}

// validateDLQ checks the arguments of a command on the dead letters.
func (f *Flags) validateDLQ() (bool, string) {
	switch {
	case f.DLQ != DLQList && f.DLQ != DLQReplay && f.DLQ != DLQPurge:
		return false, fmt.Sprintf("The command -dlq must be '%s', '%s' or '%s'. Show the help to learn more about.", DLQList, DLQReplay, DLQPurge)
	case f.IsAPIServer || f.IsWorkerServer || f.IsWebServer:
		return false, "The command -dlq starts no server: do not use -api, -worker nor -web."
	case len(f.DLQTopic) == 0:
		return false, "A topic must be defined with -dlq-topic. Show the help to learn more about."
	case f.DLQLimit < 0:
		return false, "The limit -dlq-limit must not be negative."
	case len(f.ServerName) == 0:
		return false, "A server name must be defined. Show the help to learn more about."
	case len(f.ServerConfigurationFile) == 0:
		return false, "A path to the server configuration file must be defined. Show the help to learn more about."
	}

	return true, ""
}
