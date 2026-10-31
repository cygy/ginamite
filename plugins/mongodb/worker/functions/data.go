package functions

import (
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

// CreateDefaultDocuments calls hook, which populates the database with the initial documents, with the main database.
func (f *Functions) CreateDefaultDocuments(environment, version string, hook func(string, string, store.Database)) {
	hook(environment, version, f.database())
}

// RemoveDefaultDocuments calls hook, which removes the unused documents, with the main database.
func (f *Functions) RemoveDefaultDocuments(environment, version string, hook func(string, string, store.Database)) {
	hook(environment, version, f.database())
}

// UpdateDefaultDocuments calls hook, which updates the default documents, with the main database.
func (f *Functions) UpdateDefaultDocuments(environment, version string, hook func(string, string, store.Database)) {
	hook(environment, version, f.database())
}
