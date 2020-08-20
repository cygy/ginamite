package flags

// Flags : represents the command arguments passed at launch.
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

// NewFlags : returns a newly created 'Flags' struct.
func NewFlags() *Flags {
	return &Flags{}
}
