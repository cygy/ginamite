package session

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

// Database : returns the database of the main session.
func Database() *mongo.Database {
	return Main.Database()
}

// Inject : saves the mongodb database to the context.
func Inject() gin.HandlerFunc {
	return Main.Inject()
}

// Get : returns the mongodb database saved into a request.
func Get(c *gin.Context) *mongo.Database {
	value, ok := c.Get(contextKey)
	if ok {
		return value.(*mongo.Database)
	}

	return nil
}
