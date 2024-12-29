package config

import (
	"os"
	"strings"

	"github.com/cygy/ginamite/common/config/database"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"
	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

// Initialize : initialize a configuration from a yaml file.
func (config *Configuration) Initialize(file, environment, version string) error {
	var err error
	defer func() {
		if err != nil {
			log.WithFields(logrus.Fields{
				"path":  file,
				"error": err.Error(),
			}).Panic("unable to load the configuration file")
		}
	}()

	var source []byte
	source, err = os.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(source, config)
	if err != nil {
		return err
	}

	config.Environment = environment
	config.Version = version

	for locale, c := range config.Locales {
		config.SupportedLocales = append(config.SupportedLocales, locale)

		if c.Default {
			parts := strings.Split(locale, "-")
			if len(parts) > 1 {
				language := parts[0]
				config.FallbackLocales[language] = locale
			}
		}
	}

	return nil
}

// SupportsLocale : returns true if it supports the locale.
func (config *Configuration) SupportsLocale(locale string) bool {
	lowerLocale := strings.ToLower(locale)

	var found bool
	for _, supportedLocale := range config.SupportedLocales {
		if lowerLocale == supportedLocale {
			found = true
			break
		}
	}

	return found
}

// IsTestEnvironment : returns true if the environment is Test.
func (config *Configuration) IsTestEnvironment() bool {
	return config.Environment == environment.Test
}

// IsDevelopmentEnvironment : returns true if the environment is Development.
func (config *Configuration) IsDevelopmentEnvironment() bool {
	return config.Environment == environment.Development
}

// IsProductionEnvironment : returns true if the environment is Production.
func (config *Configuration) IsProductionEnvironment() bool {
	return config.Environment == environment.Production
}

// IsDebugModeEnabled : returns true if the debug mode is enabled.
func (config *Configuration) IsDebugModeEnabled() bool {
	return config.IsDevelopmentEnvironment()
}

// IsAlphaVersion : returns true if the version is Alpha.
func (config *Configuration) IsAlphaVersion() bool {
	return config.Version == version.Alpha
}

// IsBetaVersion : returns true if the version is Beta.
func (config *Configuration) IsBetaVersion() bool {
	return config.Version == version.Beta
}

// IsFinalVersion : returns true if the version is Final.
func (config *Configuration) IsFinalVersion() bool {
	return config.Version == version.Final
}

// IsMongoDBDatabase : returns true if the database is MongoDB.
func (config *Configuration) IsMongoDBDatabase() bool {
	return config.Database.Service == database.MongoDB
}

// KafkaConfiguration : returns the kafka configuration if provided.
func (config *Configuration) KafkaConfiguration() (ServiceConfiguration, bool) {
	service, ok := config.Services["kafka"]
	return service, ok && len(service.Host) > 0 && service.Port > 0
}
