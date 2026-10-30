// Package flags parses the command line arguments of the servers.
package flags

import (
	"flag"
	"fmt"
	"io"
	"log"
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
}

// NewFlags returns an empty Flags.
func NewFlags() *Flags {
	return &Flags{}
}

// Parse parses the command line arguments of the process.
// It prints the usage and returns false if the help is requested.
// It logs a fatal error, which exits the process, if a mandatory argument is
// missing; the application configuration file is mandatory only if
// applicationConfigurationRequired is true.
func (f *Flags) Parse(applicationConfigurationRequired bool) bool {
	ok, problem := f.parse(os.Args[1:], os.Stderr, applicationConfigurationRequired)
	if len(problem) > 0 {
		log.Fatal(problem)
	}

	return ok
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

	if f.Help {
		set.PrintDefaults()
		return false, ""
	}

	switch {
	case !f.IsAPIServer && !f.IsWorkerServer && !f.IsWebServer:
		return false, "The server must be defined as 'api', 'worker' or 'web'. Show the help to learn more about."
	case len(f.ServerName) == 0:
		return false, "A server name must be defined. Show the help to learn more about."
	case len(f.ServerConfigurationFile) == 0:
		return false, "A path to the server configuration file must be defined. Show the help to learn more about."
	case applicationConfigurationRequired && len(f.ApplicationConfigurationFile) == 0:
		return false, "A path to the application configuration file must be defined. Show the help to learn more about."
	case len(f.RoutesFile) == 0:
		return false, "A path to the web routes file must be defined. Show the help to learn more about."
	}

	return true, ""
}
