// Package timezone lists the timezones and the countries of the application, loaded from a CSV file.
package timezone

import (
	"bufio"
	"os"
	"sort"
	"strings"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// Configuration holds the timezones, the countries and the links between them.
type Configuration struct {
	TimezonesByCountries map[string][]string
	CountriesByTimezones map[string][]string
	Countries            []string
	Timezones            []string
}

// NewConfiguration returns an empty Configuration.
func NewConfiguration() (conf *Configuration) {
	conf = new(Configuration)
	conf.TimezonesByCountries = map[string][]string{}
	conf.CountriesByTimezones = map[string][]string{}
	conf.Countries = []string{}
	conf.Timezones = []string{}

	return
}

const itemsByLine = 3

// GetTimezones returns the timezones of a country, given by its code (case-insensitive).
// The result is empty if the country is unknown.
func (conf *Configuration) GetTimezones(country string) []string {
	if timezones, ok := conf.TimezonesByCountries[strings.ToUpper(country)]; ok {
		return timezones
	}

	return []string{}
}

// GetCountries returns the codes of the countries of a timezone.
// The result is empty if the timezone is unknown.
func (conf *Configuration) GetCountries(timezone string) []string {
	if countries, ok := conf.CountriesByTimezones[timezone]; ok {
		return countries
	}

	return []string{}
}

// Initialize loads the countries and the timezones from a tab-separated file,
// whose lines are "country codes (comma-separated)<TAB>coordinates<TAB>timezone".
// Lines starting with "#" and malformed lines are ignored.
// It panics, after logging the error, if the file can not be read.
func (conf *Configuration) Initialize(filepath string) error {
	var err error
	defer func() {
		if err != nil {
			log.WithFields(logrus.Fields{
				"path":  filepath,
				"error": err.Error(),
			}).Panic("unable to load the timezone file")
		}
	}()

	var file *os.File
	file, err = os.Open(filepath)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.WithField("error", closeErr).Error("unable to close the timezones file")
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.SplitN(line, "\t", itemsByLine+1)
		if len(fields) < itemsByLine {
			continue
		}

		timezone := fields[itemsByLine-1]
		if !hasValue(conf.Timezones, timezone) {
			conf.Timezones = append(conf.Timezones, timezone)
		}
		if _, ok := conf.CountriesByTimezones[timezone]; !ok {
			conf.CountriesByTimezones[timezone] = []string{}
		}

		codes := strings.Split(fields[0], ",")
		for _, country := range codes {
			country = strings.ToUpper(country)
			if !hasValue(conf.Countries, country) {
				conf.Countries = append(conf.Countries, country)
			}
			if _, ok := conf.TimezonesByCountries[country]; !ok {
				conf.TimezonesByCountries[country] = []string{}
			}

			conf.CountriesByTimezones[timezone] = append(conf.CountriesByTimezones[timezone], country)
			conf.TimezonesByCountries[country] = append(conf.TimezonesByCountries[country], timezone)
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	sort.Strings(conf.Countries)
	sort.Strings(conf.Timezones)

	for country, values := range conf.TimezonesByCountries {
		sort.Strings(values)
		conf.TimezonesByCountries[country] = values
	}

	for timezone, values := range conf.CountriesByTimezones {
		sort.Strings(values)
		conf.CountriesByTimezones[timezone] = values
	}

	log.WithFields(logrus.Fields{
		"countries": len(conf.Countries),
		"timezones": len(conf.Timezones),
	}).Info("timezones loaded")

	return nil
}

// SupportsTimezone reports whether the timezone is known.
func (conf *Configuration) SupportsTimezone(timezone string) bool {
	for _, supportedTimezone := range conf.Timezones {
		if timezone == supportedTimezone {
			return true
		}
	}

	return false
}

// Load returns the configuration loaded from the file.
// See Configuration.Initialize.
func Load(filepath string) (*Configuration, error) {
	conf := NewConfiguration()
	if err := conf.Initialize(filepath); err != nil {
		return nil, err
	}

	return conf, nil
}

func hasValue(slice []string, value string) bool {
	for _, v := range slice {
		if value == v {
			return true
		}
	}

	return false
}
