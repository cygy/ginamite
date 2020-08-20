package timezone

// Initialize : initialize the main configuration of the countries and the timezones from a CSV file.
func Initialize(filepath string) error {
	Main = NewConfiguration()
	return Main.Initialize(filepath)
}
