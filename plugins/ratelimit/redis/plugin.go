// Package redis is the rate limiting plugin keeping the counters of the rate limits of the API in Redis, to share them
// between the instances of the API. Its connection is set in the subsection "redis" of the section "api_rate_limiter" of
// the configuration.
package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/plugins/ratelimit"

	goredis "github.com/redis/go-redis/v9"
)

// Store is the name of the storage of this plugin, logged when the API starts.
const Store = "redis"

// Settings is the subsection "redis" of the section "api_rate_limiter". The password can be read with !env or !file.
type Settings struct {
	Host     string
	Port     int
	Username string
	Password string

	// Database is the number of the Redis database.
	Database int

	// Prefix starts the keys of the counters.
	Prefix string

	// Timeout is the delay, in seconds, to connect to the server and to run a command.
	Timeout int
}

// DefaultSettings returns the settings with their default values.
func DefaultSettings() Settings {
	return Settings{Port: 6379, Prefix: "ginamite:ratelimit:", Timeout: 2}
}

// Plugin is the rate limiting plugin of Redis. It only works with the API: Run does not connect it for the other
// servers. Create it with NewPlugin and add it to the server.
type Plugin struct {
	settings Settings
	client   *goredis.Client
}

// NewPlugin returns a Plugin.
func NewPlugin() *Plugin {
	return &Plugin{settings: DefaultSettings()}
}

var _ ginamite.APIOnly = (*Plugin)(nil)

// APIOnly marks the plugin as needed by the API alone.
func (*Plugin) APIOnly() {}

// Connect opens the connection to Redis, with the settings of the configuration, and pings the server. It returns an
// error if the settings are invalid or the server is not reachable.
func (p *Plugin) Connect(configuration *config.Configuration) error {
	p.settings = DefaultSettings()
	if err := configuration.RateLimiterStore(Store, &p.settings); err != nil {
		return err
	}

	if p.settings.Host == "" {
		return errors.New("the setting api_rate_limiter.redis.host is required")
	}

	timeout := time.Duration(max(p.settings.Timeout, 1)) * time.Second
	client := goredis.NewClient(&goredis.Options{
		Addr:         p.settings.Host + ":" + strconv.Itoa(p.settings.Port),
		Username:     p.settings.Username,
		Password:     p.settings.Password,
		DB:           p.settings.Database,
		DialTimeout:  timeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return fmt.Errorf("rate limiter: unable to connect to Redis at %s: %w", client.Options().Addr, err)
	}
	p.client = client

	return nil
}

// Close closes the connection, if it is open.
func (p *Plugin) Close() {
	if p.client != nil {
		_ = p.client.Close()
	}
}

// Configure sets the limiters of the API. It returns an error if a rate limiting plugin is already loaded. It does
// nothing if the plugin is not connected, that is if the API does not run.
func (p *Plugin) Configure(s *ginamite.Server) error {
	if p.client == nil {
		return nil
	}

	return ratelimit.Register(s, Store, NewRateLimiters(p.client, p.settings.Prefix))
}
