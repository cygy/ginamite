package config

// Main is the global configuration of the application, set by Initialize.
var Main *Configuration

// Initialize creates the global configuration Main from a YAML file.
// See Configuration.Initialize.
func Initialize(file, environment, version string) error {
	Main = NewConfiguration()
	return Main.Initialize(file, environment, version)
}
