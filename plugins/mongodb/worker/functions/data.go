package functions

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// DefaultDocumentsHook is a function of the application run on the default documents when the worker starts, for an
// environment and a version of the application.
type DefaultDocumentsHook func(ctx context.Context, environment, version string, db store.Database)

// CreateAllDefaultDocuments populates the database with the initial documents of the plugin, for an environment and
// a version of the application. The plugin has none.
func CreateAllDefaultDocuments(_, _ string) {
}

// RemoveUnusedDefaultDocuments removes the documents of the plugin that are not used anymore, for an environment and a
// version of the application. The plugin has none.
func RemoveUnusedDefaultDocuments(_, _ string) {
}

// UpdateAllDefaultDocuments updates the default documents of the plugin, for an environment and a version of the
// application. The plugin has none.
func UpdateAllDefaultDocuments(_, _ string) {
}

// CreateDefaultDocuments calls hook, which populates the database with the initial documents, with the main database.
func (f *Functions) CreateDefaultDocuments(ctx context.Context, environment, version string, hook DefaultDocumentsHook) {
	hook(ctx, environment, version, f.database())
}

// RemoveDefaultDocuments calls hook, which removes the unused documents, with the main database.
func (f *Functions) RemoveDefaultDocuments(ctx context.Context, environment, version string, hook DefaultDocumentsHook) {
	hook(ctx, environment, version, f.database())
}

// UpdateDefaultDocuments calls hook, which updates the default documents, with the main database.
func (f *Functions) UpdateDefaultDocuments(ctx context.Context, environment, version string, hook DefaultDocumentsHook) {
	hook(ctx, environment, version, f.database())
}
