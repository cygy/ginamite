package timezone

// Configuration : all the timezones configuration.
type Configuration struct {
	TimezonesByCountries map[string][]string
	CountriesByTimezones map[string][]string
	Countries            []string
	Timezones            []string
}

// NewConfiguration : returns a new struct Configuration.
func NewConfiguration() (conf *Configuration) {
	conf = new(Configuration)
	conf.TimezonesByCountries = map[string][]string{}
	conf.CountriesByTimezones = map[string][]string{}
	conf.Countries = []string{}
	conf.Timezones = []string{}

	return
}
