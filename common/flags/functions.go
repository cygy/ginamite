package flags

import (
	"flag"
	"fmt"
	"log"

	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"
)

// Parse : parses the flags passed at launch.
func (f *Flags) Parse(applicationConfigurationRequired bool) bool {
	isAPIServer := flag.Bool("api", false, "Present if the server is a API server.")
	isWorkerServer := flag.Bool("worker", false, "Present if the server is a worker server.")
	isWebServer := flag.Bool("web", false, "Present if the server is a web server.")
	serverName := flag.String("name", "", "Name of the server.")
	env := flag.String("env", environment.Development, fmt.Sprintf("Running environment of the server (%s, %s or %s)", environment.Development, environment.Production, environment.Test))
	version := flag.String("version", version.Final, fmt.Sprintf("Running version of the server (%s, %s or %s)", version.Final, version.Beta, version.Alpha))
	serverConfigurationFile := flag.String("conf-server", "", "Path to the server configuration file.")
	applicationConfigurationFile := flag.String("conf-app", "", "Path to the application configuration file.")
	routesFile := flag.String("routes", "", "Path to the web routes file.")
	help := flag.Bool("help", false, "Print the help message")
	flag.Parse()

	f.IsAPIServer = *isAPIServer
	f.IsWorkerServer = *isWorkerServer
	f.IsWebServer = *isWebServer
	f.ServerName = *serverName
	f.Environment = *env
	f.Version = *version
	f.ServerConfigurationFile = *serverConfigurationFile
	f.ApplicationConfigurationFile = *applicationConfigurationFile
	f.RoutesFile = *routesFile
	f.Help = *help

	// Print the help message.
	if f.Help {
		flag.PrintDefaults()
		return false
	}

	// Verify that a API server or a worker is defined.
	if !f.IsAPIServer && !f.IsWorkerServer && !f.IsWebServer {
		log.Fatal("The server must be defined as 'api', 'worker' or 'web'. Show the help to learn more about.")
		return false
	}

	// Verify that a server name is provided.
	if len(f.ServerName) == 0 {
		log.Fatal("A server name must be defined. Show the help to learn more about.")
		return false
	}

	// Verify that a path to the configuration files are defined.
	if len(f.ServerConfigurationFile) == 0 {
		log.Fatal("A path to the server configuration file must be defined. Show the help to learn more about.")
		return false
	}
	if applicationConfigurationRequired && len(f.ApplicationConfigurationFile) == 0 {
		log.Fatal("A path to the application configuration file must be defined. Show the help to learn more about.")
		return false
	}
	if len(f.RoutesFile) == 0 {
		log.Fatal("A path to the web routes file must be defined. Show the help to learn more about.")
		return false
	}

	return true
}
