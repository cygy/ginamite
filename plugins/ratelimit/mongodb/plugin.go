// Package mongodb is the rate limiting plugin keeping the counters of the rate limits of the API in MongoDB, to share
// them between the instances of the API. It has its own connection, set in the subsection "mongodb" of the section
// "api_rate_limiter" of the configuration, and does not need the MongoDB plugin of the data of the application.
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/ratelimit"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Store is the name of the storage of this plugin, logged when the API starts.
const Store = "mongodb"

// DefaultCollection is the default name of the collection of the counters. A document counts the requests of a client
// in a window of time and expires when the window ends.
const DefaultCollection = "rateLimits"

// errorCodeNamespaceExists is the code of the error of MongoDB for a collection which already exists.
const errorCodeNamespaceExists = 48

// Settings is the subsection "mongodb" of the section "api_rate_limiter". The password can be read with !env or !file.
type Settings struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string

	// Timeout is the delay, in seconds, to connect to the server.
	Timeout int

	// Collection is the name of the collection of the counters.
	Collection string
}

// DefaultSettings returns the settings with their default values.
func DefaultSettings() Settings {
	return Settings{Port: 27017, Timeout: 10, Collection: DefaultCollection}
}

// Plugin is the rate limiting plugin of MongoDB. It only works with the API: Run does not connect it for the other
// servers. Create it with NewPlugin and add it to the server.
type Plugin struct {
	settings Settings
	session  *session.Session
	database store.Database // The database of the counters, set by Connect.
}

// NewPlugin returns a Plugin.
func NewPlugin() *Plugin {
	return &Plugin{settings: DefaultSettings()}
}

var _ ginamite.APIOnly = (*Plugin)(nil)

// APIOnly marks the plugin as needed by the API alone.
func (*Plugin) APIOnly() {}

// Connect opens the connection to the database of the counters, with the settings of the configuration. It returns an
// error if the settings are invalid or the server is not reachable.
func (p *Plugin) Connect(configuration *config.Configuration) error {
	p.settings = DefaultSettings()
	if err := configuration.RateLimiterStore(Store, &p.settings); err != nil {
		return err
	}

	if p.settings.Host == "" || p.settings.Database == "" {
		return errors.New("the settings api_rate_limiter.mongodb.host and api_rate_limiter.mongodb.database are required")
	}

	s := session.NewSession()
	if err := s.Initialize(p.settings.Host+":"+strconv.Itoa(p.settings.Port),
		p.settings.Database,
		p.settings.Username,
		p.settings.Password,
		time.Duration(p.settings.Timeout)); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}
	p.session, p.database = s, s.Database()

	return nil
}

// Close closes the connection, if it is open.
func (p *Plugin) Close() {
	if p.session != nil {
		p.session.Close()
	}
}

// Configure creates the collection of the counters and its expiration index, then sets the limiters of the API. It
// returns an error if a rate limiting plugin is already loaded. It does nothing if the plugin is not connected, that is
// if the API does not run.
func (p *Plugin) Configure(s *ginamite.Server) error {
	if p.database == nil {
		return nil
	}

	if err := prepare(context.Background(), p.database, p.settings.Collection); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	return ratelimit.Register(s, Store, NewRateLimiters(func() store.Database { return p.database }, p.settings.Collection))
}

// prepare creates the collection and the index removing the counters when their window ends. It is idempotent.
func prepare(ctx context.Context, database store.Database, collection string) error {
	err := database.CreateCollection(ctx, collection)

	var serverError mongo.ServerError
	if err != nil && !(errors.As(err, &serverError) && serverError.HasErrorCode(errorCodeNamespaceExists)) {
		return fmt.Errorf("unable to create the collection %s: %w", collection, err)
	}

	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "expireAt", Value: 1}},
		Options: options.Index().SetName("expireAt_1_1").SetExpireAfterSeconds(0),
	}
	if _, err := database.Collection(collection).Indexes().CreateOne(ctx, index); err != nil {
		return fmt.Errorf("unable to create the index of the collection %s: %w", collection, err)
	}

	return nil
}
