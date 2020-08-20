package config

// Initialize : initialize the global configuration of the application.
func Initialize(file, environment, version string) error {
	Main = NewConfiguration()
	return Main.Initialize(file, environment, version)
}
