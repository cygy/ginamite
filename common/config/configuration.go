package config

import (
	"os"
	"strings"

	"github.com/cygy/ginamite/common/config/database"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/worker/tasks/recurring"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

// Configuration is the root configuration object.
type Configuration struct {
	Environment         string
	Version             string
	SupportedLocales    []string                       // Filled by Initialize.
	FallbackLocales     map[string]string              // Maps a language ("fr", "en", ...) to its default locale ("fr-fr", "en-gb", ...). Filled by Initialize.
	Locales             map[string]LocaleConfiguration `yaml:"locales"`
	DefaultLocale       string                         `yaml:"default_locale"`
	DefaultTimezone     string                         `yaml:"default_timezone"`
	LocalizationPath    string                         `yaml:"localization_path"`
	TemplatesPath       string                         `yaml:"templates_path"`
	RobotsFilePath      string                         `yaml:"robots_file_path"`
	TimezonesFilePath   string                         `yaml:"timezones_file_path"`
	APIVersion          string                         `yaml:"api_version"`
	SenderName          string                         `yaml:"sender_name"`
	SenderEmailAddress  string                         `yaml:"sender_email_address"`
	NoReplyEmailAddress string                         `yaml:"noreply_email_address"`
	Server              ServerConfiguration
	Hosts               HostConfiguration
	Database            DatabaseConfiguration
	Cookie              CookieConfiguration
	JWT                 JWTConfiguration  `yaml:"jwt"`
	SMTP                SMTPConfiguration `yaml:"smtp"`
	API                 APIConfiguration  `yaml:"api"`
	Facebook            FacebookConfiguration
	Google              GoogleConfiguration
	SocialNetworks      SocialMediaConfiguration       `yaml:"social_networks"`
	IPStack             IPStackConfiguration           `yaml:"ipstack"`
	ExpiredAuthTokens   ExpiredAuthTokensConfiguration `yaml:"expired_auth_tokens"`
	Sitemap             SitemapConfiguration
	Terms               TermsConfiguration
	UserNotifications   UserNotificationsConfiguration `yaml:"user_notifications"`
	Account             AccountConfiguration
	Metrics             MetricsConfiguration      `yaml:"metrics"`          // Prometheus metrics, disabled by default
	RateLimits          RateLimitsConfiguration   `yaml:"rate_limits"`      // limits of the requests by client, 0 to disable
	Broker              BrokerConfiguration       `yaml:"broker"`           // address of the message broker, and the plugin which connects to it
	Kafka               KafkaConfiguration        `yaml:"kafka"`            // connection to Kafka, its topics and its retries
	WorkerConsumers     []ConsumerConfiguration   `yaml:"worker_consumers"` // topics consumed by the worker, all if empty
	RecurringTasks      []recurring.ScheduledTask `yaml:"recurring_tasks"`  // tasks executed by the workers
	RunStartUpTasks     bool                      `yaml:"run_start_up_tasks"`
}

// NewConfiguration returns a Configuration filled with the default values.
func NewConfiguration() (config *Configuration) {
	config = new(Configuration)
	config.Environment = environment.Development
	config.Version = version.Beta
	config.Server.Host = "localhost"
	config.Server.TrustedProxies = []string{}
	config.Server.Port = 80
	config.Server.GZip.Enabled = false
	config.Server.SSL.Enabled = false
	config.DefaultLocale = "en-us"
	config.Locales = map[string]LocaleConfiguration{}
	config.SupportedLocales = []string{}
	config.FallbackLocales = map[string]string{}
	config.Database.TimeOut = 60
	config.Cookie.Secure = true
	config.Cookie.HTTPOnly = true
	config.JWT.SigningMethod = "HS256"
	config.JWT.TTL = 3600
	config.API.TimeOut = 2000
	config.RateLimits.Authentication = RateLimitConfiguration{Requests: 10, Seconds: 60}
	config.RateLimits.Registration = RateLimitConfiguration{Requests: 5, Seconds: 60}
	config.RateLimits.ForgotPassword = RateLimitConfiguration{Requests: 5, Seconds: 60}
	config.Kafka.MaxAttempts = 3
	config.Kafka.RetryDelay = 1000
	config.Kafka.Topics.Partitions = 1
	config.Kafka.Topics.ReplicationFactor = 1
	config.Kafka.Topics.RetentionHours = 168
	config.Kafka.Topics.DeadLetterRetentionHours = 720
	config.Facebook.Enabled = false
	config.Google.Enabled = false
	config.ExpiredAuthTokens.FlushCacheInterval = 60
	config.Sitemap.Name = "sitemap"
	config.Sitemap.MaximumURLsPerPage = 50000
	config.Terms.Version = "0"
	config.UserNotifications.Limit = 20
	config.Account.EmailAddressMustBeConfirmed = true
	config.Account.DeleteNeverUsedAfter = 10
	config.Account.DeleteInactiveAfter = 48
	config.Account.InvalidStringsForUsername = []string{}
	config.RecurringTasks = []recurring.ScheduledTask{}
	config.RunStartUpTasks = false

	return
}

// Initialize loads the configuration from the YAML file, sets its environment and version,
// and computes the supported and fallback locales.
// It panics, after logging the error, if the file can not be read or parsed.
func (config *Configuration) Initialize(file, env, ver string) error {
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

	config.Environment = env
	config.Version = ver

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

// SupportsLocale reports whether the locale, compared case-insensitively, is supported.
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

// IsTestEnvironment reports whether the environment is Test.
func (config *Configuration) IsTestEnvironment() bool {
	return config.Environment == environment.Test
}

// IsDevelopmentEnvironment reports whether the environment is Development.
func (config *Configuration) IsDevelopmentEnvironment() bool {
	return config.Environment == environment.Development
}

// IsProductionEnvironment reports whether the environment is Production.
func (config *Configuration) IsProductionEnvironment() bool {
	return config.Environment == environment.Production
}

// IsDebugModeEnabled reports whether the debug mode is enabled, i.e. in the Development environment.
func (config *Configuration) IsDebugModeEnabled() bool {
	return config.IsDevelopmentEnvironment()
}

// IsAlphaVersion reports whether the version is Alpha.
func (config *Configuration) IsAlphaVersion() bool {
	return config.Version == version.Alpha
}

// IsBetaVersion reports whether the version is Beta.
func (config *Configuration) IsBetaVersion() bool {
	return config.Version == version.Beta
}

// IsFinalVersion reports whether the version is Final.
func (config *Configuration) IsFinalVersion() bool {
	return config.Version == version.Final
}

// IsMongoDBDatabase reports whether the database is MongoDB.
func (config *Configuration) IsMongoDBDatabase() bool {
	return config.Database.Service == database.MongoDB
}
