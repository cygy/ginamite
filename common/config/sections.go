// Package config loads and exposes the configuration of the application.
package config

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.yaml.in/yaml/v3"
)

// GZipConfiguration holds the GZip settings.
type GZipConfiguration struct {
	Enabled bool
	Level   int
}

// SSLConfiguration holds the SSL settings.
type SSLConfiguration struct {
	Enabled  bool
	CertPath string `yaml:"cert_path"`
	KeyPath  string `yaml:"key_path"`
}

// ServerConfiguration holds the settings of the HTTP server.
type ServerConfiguration struct {
	Host string
	Port int
	// TrustedProxies are the addresses or the networks of the proxies allowed to give the address of the client in the
	// headers X-User-IP, X-Forwarded-For and X-Real-IP, such as ["10.0.0.0/8"]: the reverse proxies and, for the API,
	// the web servers. No proxy is trusted by default: the client address, used to limit the requests and saved with
	// the auth tokens, is the one of the connection.
	TrustedProxies []string `yaml:"trusted_proxies"`
	// MaxBodySize is the maximum size of the body of a request, in bytes: 10 MiB by default, which must hold the
	// largest uploaded file. Larger bodies are rejected with the status 413. Zero disables the limit.
	MaxBodySize int64             `yaml:"max_body_size"`
	GZip        GZipConfiguration `yaml:"gzip"`
	SSL         SSLConfiguration  `yaml:"ssl"`
	LogFormat   string            `yaml:"log_format"`
}

// APIConfiguration holds the settings of an API called by the server.
type APIConfiguration struct {
	Host       string
	Port       int
	Version    string
	TimeOut    int `yaml:"time_out"` // in milliseconds
	RetryCount int `yaml:"retry_count"`
}

// HostConfiguration holds the hosts of the application.
type HostConfiguration struct {
	API    string `yaml:"api"`
	Web    string
	Static string
}

// LocaleConfiguration holds the formats of a locale.
type LocaleConfiguration struct {
	DateFormat     string `yaml:"date_format"`
	TimeFormat     string `yaml:"time_format"`
	DateTimeFormat string `yaml:"date_time_format"`
	Default        bool   `yaml:"default"`
}

// DatabaseConfiguration holds the settings of the database.
type DatabaseConfiguration struct {
	Service  string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	TimeOut  int64 `yaml:"time_out"`
}

// CookieConfiguration holds the settings of the cookie.
type CookieConfiguration struct {
	Name     string
	Domain   string
	Path     string
	TTL      int `yaml:"ttl"`
	Secure   bool
	HTTPOnly bool `yaml:"http_only"`
	// SameSite is the SameSite attribute of the cookie: lax (default), strict or none, which requires secure.
	SameSite string `yaml:"same_site"`
}

// SameSiteMode returns the SameSite attribute of the cookie, http.SameSiteLaxMode if the setting is unknown.
func (c CookieConfiguration) SameSiteMode() http.SameSite {
	switch strings.ToLower(c.SameSite) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// validate returns an error if the SameSite attribute is unknown, or none without secure: the browsers reject such a
// cookie.
func (c CookieConfiguration) validate() error {
	switch strings.ToLower(c.SameSite) {
	case "lax", "strict":
		return nil
	case "none":
		if !c.Secure {
			return errors.New("invalid setting 'cookie.same_site': none requires 'cookie.secure'")
		}
		return nil
	default:
		return fmt.Errorf("invalid setting 'cookie.same_site' %q: lax, strict or none expected", c.SameSite)
	}
}

// JWTConfiguration holds the settings of the JWT tokens.
type JWTConfiguration struct {
	Secret        string
	SigningMethod string `yaml:"signing_method"`
	TTL           int64  `yaml:"ttl"`
}

// SMTPConfiguration holds the settings of the SMTP server.
type SMTPConfiguration struct {
	Host     string
	Port     int
	Username string
	Password string
}

// ConsumerConfiguration holds the settings of the consumption of a Kafka topic.
type ConsumerConfiguration struct {
	Topic string
	// Group is the consumer group of the workers. The workers of the same group share the messages of the topic, the
	// workers of different groups each receive every message. It is the topic if empty.
	Group string
}

// FacebookConfiguration holds the settings of the Facebook developer application.
type FacebookConfiguration struct {
	Enabled bool
	AppID   string `yaml:"app_id"`
	// AppSecret is the secret of the application, which verifies the tokens of the users: read it with !file or !env.
	AppSecret  string `yaml:"app_secret"`
	APIVersion string `yaml:"api_version"`
	TimeOut    int    `yaml:"time_out"` // in milliseconds
	RetryCount int    `yaml:"retry_count"`
}

// GoogleConfiguration holds the settings of the Google developer application.
type GoogleConfiguration struct {
	Enabled    bool
	ClientID   string `yaml:"client_id"`
	APIVersion string `yaml:"api_version"`
	TimeOut    int    `yaml:"time_out"` // in milliseconds
	RetryCount int    `yaml:"retry_count"`
}

// SocialMediaConfiguration holds the URLs of the social network pages.
type SocialMediaConfiguration struct {
	Facebook string
	Google   string
	Twitter  string
}

// IPStackConfiguration holds the settings of the ipstack.com API.
type IPStackConfiguration struct {
	Host string
	Key  string
}

// ExpiredAuthTokensConfiguration holds the settings of the management of the expired auth tokens.
type ExpiredAuthTokensConfiguration struct {
	FlushCacheInterval int `yaml:"flush_cache_interval"`
}

// TermsConfiguration holds the settings of the terms of use of the application.
type TermsConfiguration struct {
	Path    string `yaml:"path"`
	Version string `yaml:"version"`
}

// UserNotificationsConfiguration holds the settings of the notifications received by a user.
type UserNotificationsConfiguration struct {
	Limit int64 `yaml:"limit"`
}

// AccountConfiguration holds the settings of the users' accounts.
type AccountConfiguration struct {
	EmailAddressMustBeConfirmed bool     `yaml:"email_address_must_be_confirmed"`
	DeleteNeverUsedAfter        uint     `yaml:"delete_never_used_after"` // in days
	DeleteInactiveAfter         uint     `yaml:"delete_inactive_after"`   // in months
	InvalidStringsForUsername   []string `yaml:"invalid_strings_for_username"`
}

// RecurringTask is a task run by the workers the first time at StartAt, then every Interval seconds.
type RecurringTask struct {
	Name     string `yaml:"name"`
	Interval uint   `yaml:"interval"`
	StartAt  string `yaml:"start_at"` // An hour in UTC, i.e. "14:30" or "4:15am".
	Enabled  bool   `yaml:"enabled"`
}

// SitemapConfiguration holds the settings of the generation of the sitemap files.
type SitemapConfiguration struct {
	Path               string `yaml:"path"`
	Name               string `yaml:"name"`
	MaximumURLsPerPage int    `yaml:"max_urls_per_page"`
}

// RateLimitConfiguration limits the requests of a client: at most Requests in Seconds seconds, by IP address. The
// requests are not limited if Requests is 0.
type RateLimitConfiguration struct {
	Requests int
	Seconds  int
}

// RateLimitsConfiguration holds the limits of the requests which try to guess a secret, by group of routes. They apply
// whatever the storage of the counters is, see APIRateLimiterConfiguration.
type RateLimitsConfiguration struct {
	Authentication RateLimitConfiguration `yaml:"authentication"`
	Registration   RateLimitConfiguration `yaml:"registration"`
	ForgotPassword RateLimitConfiguration `yaml:"forgot_password"`
}

// APIRateLimiterConfiguration is the section "api_rate_limiter": the limits of the requests, and the settings of the
// storage of the counters. By default the counters are kept in memory by each instance of the API; a rate limiting
// plugin, see api.Server.Hooks.NewRateLimiter, keeps them elsewhere and reads its own subsection, such as "mongodb" or
// "redis", with Configuration.RateLimiterStore.
type APIRateLimiterConfiguration struct {
	Limits RateLimitsConfiguration `yaml:"limits"`

	// Stores are the subsections of the rate limiting plugins, by name: read them with Configuration.RateLimiterStore.
	Stores map[string]yaml.Node `yaml:",inline"`
}

// MetricsConfiguration holds the settings of the Prometheus metrics.
type MetricsConfiguration struct {
	// Address is the address (host:port) the server listens to, on /metrics, for Prometheus. It should not be
	// reachable from the Internet. The metrics are disabled if empty.
	Address string `yaml:"address"`
}

// BrokerConfiguration holds the address of the message broker and the plugin which connects to it.
type BrokerConfiguration struct {
	// Service is the name of the broker, such as "kafka": the plugin of this broker is the one which connects. No
	// message is sent if it is empty.
	Service string
	Host    string
	Port    int
}
