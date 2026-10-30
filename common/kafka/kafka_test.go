package kafka

import (
	"os"
	"testing"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// hook collects the entries of the logger. The logger is never swapped back
// because the goroutines started by Initialize keep logging after the tests.
var hook *logtest.Hook

func TestMain(m *testing.M) {
	var logger *logrus.Logger
	logger, hook = logtest.NewNullLogger()
	log.Main = logger.WithField("server", "test")

	os.Exit(m.Run())
}

// useHook empties the hook and returns it.
func useHook(t *testing.T) *logtest.Hook {
	t.Helper()

	hook.Reset()

	return hook
}

func TestRecover(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		hook := useHook(t)

		func() {
			defer Recover()
			panic("boom")
		}()

		entry := hook.LastEntry()
		if entry == nil || entry.Level != logrus.WarnLevel || entry.Data["panic"] != "boom" {
			t.Errorf("entry = %+v", entry)
		}
	})

	t.Run("no panic", func(t *testing.T) {
		hook := useHook(t)

		func() { defer Recover() }()

		for _, e := range hook.AllEntries() {
			if e.Message == "recover from panic" {
				t.Error("nothing must be logged")
			}
		}
	})
}

func TestSendMessageToTopic(t *testing.T) {
	defer func(old *Client) { Main = old }(Main)

	hook := useHook(t)
	Main = NewClient()

	SendMessageToTopic("hello", "topic")

	entry := hook.LastEntry()
	if entry == nil || entry.Level != logrus.ErrorLevel || entry.Data["error"] != "not a producer" {
		t.Errorf("entry = %+v", entry)
	}
}
