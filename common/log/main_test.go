package log

import (
	"log/slog"
	"os"
	"testing"

	"github.com/cygy/ginamite/common/log/logtest"
)

func TestMain(m *testing.M) {
	Initialize("test", false)
	os.Exit(m.Run())
}

// useHook replaces the main logger with a silent one and returns the hook collecting its entries.
func useHook(t *testing.T) *logtest.Hook {
	t.Helper()

	hook, restore := logtest.Replace(&Main, slog.LevelDebug)
	t.Cleanup(restore)

	return hook
}
