// Package session manages the connection to the MongoDB database.
package session

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/store"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Session is a session of a MongoDB database.
// The client is safe for concurrent use and manages its own connection pool.
type Session struct {
	Client       *mongo.Client
	DatabaseName string

	// Store, if set, is returned by Database instead of the database of the client. It lets the tests
	// use a stub, see the mock package of the store.
	Store store.Database
}

// NewSession returns an empty Session.
func NewSession() (session *Session) {
	session = &Session{}
	return
}

const contextKey = "_mgo"

// Initialize connects the session to the database and checks that the server is reachable.
// The timeout is in seconds, and defaults to 10 seconds if not positive. The user, if any, is
// authenticated against the database itself.
// It panics, after logging the error, if the server can not be reached.
func (session *Session) Initialize(address, database, username, password string, timeout time.Duration) {
	session.DatabaseName = database

	timeout = timeout * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	opts := options.Client().
		SetHosts([]string{address}).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout)

	// Like with mgo, the user is authenticated against the database itself.
	if username != "" {
		opts.SetAuth(options.Credential{
			Username:   username,
			Password:   password,
			AuthSource: database,
		})
	}

	// mongo.Connect does not wait for the connection: the ping checks that the server is reachable.
	client, err := mongo.Connect(opts)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err = client.Ping(ctx, readpref.Primary()); err != nil {
			_ = client.Disconnect(context.Background())
		}
	}

	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Panic("unable to connect to the mongo database")
	}

	session.Client = client
}

// Close disconnects the session. It does nothing if the session is not connected.
func (session *Session) Close() {
	if session.Client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := session.Client.Disconnect(ctx); err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("unable to close the mongo connection")
	}
}

// Database returns the database of the session.
// There is no need to copy the session or to close the database: the client manages a pool of connections.
func (session *Session) Database() store.Database {
	if session.Store != nil {
		return session.Store
	}

	return store.NewDatabase(session.Client.Database(session.DatabaseName))
}

// Inject returns a middleware saving the database of the session in the Gin context, see Get.
func (session *Session) Inject() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(contextKey, session.Database())
		c.Next()
	}
}

// Set saves the database in the Gin context, like Inject does. It is meant for the tests, to use a stub.
func Set(c *gin.Context, db store.Database) {
	c.Set(contextKey, db)
}

// Get returns the database saved in the Gin context by Inject, or nil if there is none.
func Get(c *gin.Context) store.Database {
	value, ok := c.Get(contextKey)
	if ok {
		return value.(store.Database)
	}

	return nil
}
