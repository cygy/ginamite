// Package functions provides the functions of the workers working on the MongoDB database of the plugin.
package functions

import "github.com/cygy/ginamite/plugins/mongodb/mongo/store"

// Functions are the functions of the workers working on a database.
// Create it with New.
type Functions struct {
	database func() store.Database
}

// New returns the functions working on the database returned by the function, which is called each time a function
// needs the database.
func New(database func() store.Database) *Functions {
	return &Functions{database: database}
}
