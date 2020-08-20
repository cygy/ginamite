package session

import (
	"time"

	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/sirupsen/logrus"
)

const contextKey = "_mgo"

// Initialize : initializes the global mongodb session.
func (session *Session) Initialize(address, database, username, password string, timeout time.Duration) {
	session.DatabaseName = database
	databaseInfo := &mgo.DialInfo{
		Addrs:    []string{address},
		Timeout:  timeout * time.Second,
		Database: database,
		Username: username,
		Password: password,
	}

	var err error
	session.Session, err = mgo.DialWithInfo(databaseInfo)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Panic("unable to connect to the mongo database")
	}
}

// Close : closes the session.
func (session *Session) Close() {
	if session.Session != nil {
		session.Session.Close()
	}
}

// Copy : copies the session.
func (session *Session) Copy() (*mgo.Session, *mgo.Database) {
	mongoSession := session.Session.Copy()
	database := mongoSession.DB(session.DatabaseName)

	return mongoSession, database
}

// Inject : saves a mongodb session to the context.
func (session *Session) Inject() gin.HandlerFunc {
	return func(c *gin.Context) {
		mongoSession, database := session.Copy()
		defer mongoSession.Close()

		c.Set(contextKey, database)

		c.Next()
	}
}
