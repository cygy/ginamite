// Package config loads and exposes the configuration of the application.
package config

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
	// header X-Forwarded-For, such as ["10.0.0.0/8"]. Without this setting every proxy is trusted, and the client address,
	// used to limit the requests, can be forged: set it, even to an empty list if there is no proxy.
	TrustedProxies []string          `yaml:"trusted_proxies"`
	GZip           GZipConfiguration `yaml:"gzip"`
	SSL            SSLConfiguration  `yaml:"ssl"`
	LogFormat      string            `yaml:"log_format"`
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

// ServiceConfiguration holds the address of a hosted service.
type ServiceConfiguration struct {
	Host string
	Port int
}

// FacebookConfiguration holds the settings of the Facebook developer application.
type FacebookConfiguration struct {
	Enabled    bool
	AppID      string `yaml:"app_id"`
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

// RateLimitsConfiguration holds the limits of the requests which try to guess a secret, by group of routes. The counters
// are kept in memory by each instance of the API.
type RateLimitsConfiguration struct {
	Authentication RateLimitConfiguration `yaml:"authentication"`
	Registration   RateLimitConfiguration `yaml:"registration"`
	ForgotPassword RateLimitConfiguration `yaml:"forgot_password"`
}
