package timezone

import (
	"bufio"
	"os"
	"sort"
	"strings"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// GetTimezones : returns all the timezones of a country.
func (conf *Configuration) GetTimezones(country string) []string {
	if timezones, ok := conf.TimezonesByCountries[strings.ToUpper(country)]; ok {
		return timezones
	}

	return []string{}
}

// GetCountries : returns all the countries of a timezone.
func (conf *Configuration) GetCountries(timezone string) []string {
	if countries, ok := conf.CountriesByTimezones[timezone]; ok {
		return countries
	}

	return []string{}
}

// Initialize : initialize the countries and the timezones from a CSV file.
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

	defer file.Close()

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

// SupportsTimezone : returns true if it supports the timezone.
func (conf *Configuration) SupportsTimezone(timezone string) bool {
	for _, supportedTimezone := range conf.Timezones {
		if timezone == supportedTimezone {
			return true
		}
	}

	return false
}
