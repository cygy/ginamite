// Package flags parses the command line arguments of the servers.
package flags

import (
	"errors"
	"flag"
	"fmt"
	"io"

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
}

// NewFlags returns an empty Flags.
func NewFlags() *Flags {
	return &Flags{}
}

// Parse parses the command line arguments args. The plugins add their own flags with the functions register, which
// receive the flag set before it parses. It returns flag.ErrHelp, after printing the usage to output, if the help is
// requested, and an error if an argument is unknown. It does not check that the arguments describe a server to run,
// see Validate: a plugin may run a command instead.
func (f *Flags) Parse(args []string, output io.Writer, register ...func(*flag.FlagSet)) error {
	set := flag.NewFlagSet("server", flag.ContinueOnError)
	set.SetOutput(output)

	set.BoolVar(&f.IsAPIServer, "api", false, "Present if the server is a API server.")
	set.BoolVar(&f.IsWorkerServer, "worker", false, "Present if the server is a worker server.")
	set.BoolVar(&f.IsWebServer, "web", false, "Present if the server is a web server.")
	set.StringVar(&f.ServerName, "name", "", "Name of the server.")
	set.StringVar(&f.Environment, "env", environment.Development, fmt.Sprintf("Running environment of the server (%s, %s or %s)", environment.Development, environment.Production, environment.Test))
	set.StringVar(&f.Version, "version", version.Final, fmt.Sprintf("Running version of the server (%s, %s or %s)", version.Final, version.Beta, version.Alpha))
	set.StringVar(&f.ServerConfigurationFile, "conf-server", "", "Path to the server configuration file.")
	set.StringVar(&f.ApplicationConfigurationFile, "conf-app", "", "Path to the application configuration file.")
	set.StringVar(&f.RoutesFile, "routes", "", "Path to the web routes file.")

	for _, add := range register {
		add(set)
	}

	return set.Parse(args)
}

// Validate checks that the arguments describe a server to run: its type, its name and its configuration files. The
// application configuration file is mandatory only if applicationConfigurationRequired is true.
func (f *Flags) Validate(applicationConfigurationRequired bool) error {
	switch {
	case !f.IsAPIServer && !f.IsWorkerServer && !f.IsWebServer:
		return errors.New("the server must be defined as 'api', 'worker' or 'web': show the help to learn more about")
	case len(f.ServerName) == 0:
		return errors.New("a server name must be defined: show the help to learn more about")
	case len(f.ServerConfigurationFile) == 0:
		return errors.New("a path to the server configuration file must be defined: show the help to learn more about")
	case applicationConfigurationRequired && len(f.ApplicationConfigurationFile) == 0:
		return errors.New("a path to the application configuration file must be defined: show the help to learn more about")
	case f.IsWebServer && len(f.RoutesFile) == 0:
		return errors.New("a path to the web routes file must be defined: show the help to learn more about")
	}

	return nil
}
