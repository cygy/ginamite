package session

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const contextKey = "_mgo"

// Initialize : initializes the global mongodb session.
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

// Close : closes the session.
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

// Database : returns the database of the session.
// There is no need to copy the session or to close the database: the client manages a pool of connections.
func (session *Session) Database() *mongo.Database {
	return session.Client.Database(session.DatabaseName)
}

// Inject : saves the mongodb database to the context.
func (session *Session) Inject() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(contextKey, session.Database())
		c.Next()
	}
}
