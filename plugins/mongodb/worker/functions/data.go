package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store"
)

// CreateAllDefaultDocuments populates the database with the initial documents of the plugin, for an environment and
// a version of the application. The plugin has none.
func CreateAllDefaultDocuments(environment, version string) {
}

// RemoveUnusedDefaultDocuments removes the documents of the plugin that are not used anymore, for an environment and a
// version of the application. The plugin has none.
func RemoveUnusedDefaultDocuments(environment, version string) {
}

// UpdateAllDefaultDocuments updates the default documents of the plugin, for an environment and a version of the
// application. The plugin has none.
func UpdateAllDefaultDocuments(environment, version string) {
}

// CreateDefaultDocuments calls f, which populates the database with the initial documents, with the main database.
func CreateDefaultDocuments(environment, version string, f func(string, string, store.Database)) {
	db := session.Database()
	f(environment, version, db)
}

// RemoveDefaultDocuments calls f, which removes the unused documents, with the main database.
func RemoveDefaultDocuments(environment, version string, f func(string, string, store.Database)) {
	db := session.Database()
	f(environment, version, db)
}

// UpdateDefaultDocuments calls f, which updates the default documents, with the main database.
func UpdateDefaultDocuments(environment, version string, f func(string, string, store.Database)) {
	db := session.Database()
	f(environment, version, db)
}
