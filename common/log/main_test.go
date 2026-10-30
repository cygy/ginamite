package log

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

func TestMain(m *testing.M) {
	Initialize("test", false)
	os.Exit(m.Run())
}

// useHook replaces the main logger with a silent one and returns the hook collecting its entries.
func useHook(t *testing.T) *logtest.Hook {
	t.Helper()

	logger, hook := logtest.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	logger.ExitFunc = func(int) {}

	old := Main
	Main = logger.WithField("server", "test")
	t.Cleanup(func() { Main = old })

	return hook
}
