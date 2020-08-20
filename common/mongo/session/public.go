package session

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

// Initialize : initializes the main mongodb session.
func Initialize(address, database, username, password string, timeout time.Duration) {
	Main = NewSession()
	Main.Initialize(address, database, username, password, timeout)
}

// Close : closes the main session.
func Close() {
	Main.Close()
}

// Copy : copies the main session.
func Copy() (*mgo.Session, *mgo.Database) {
	return Main.Copy()
}

// Inject : saves a mongodb session to the context.
func Inject() gin.HandlerFunc {
	return Main.Inject()
}

// Get : returns the mongodb session saved into a request.
func Get(c *gin.Context) *mgo.Database {
	value, ok := c.Get(contextKey)
	if ok {
		return value.(*mgo.Database)
	}

	return nil
}
