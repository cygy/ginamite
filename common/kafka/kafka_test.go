package kafka

import (
	"errors"
	"fmt"
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

func TestPermanent(t *testing.T) {
	boom := errors.New("boom")

	if Permanent(nil) != nil {
		t.Error("Permanent(nil) must be nil")
	}

	err := Permanent(boom)
	if !IsPermanent(err) || !errors.Is(err, boom) || err.Error() != "boom" {
		t.Errorf("Permanent() = %v", err)
	}
	if !IsPermanent(fmt.Errorf("wrapped: %w", err)) {
		t.Error("a wrapped permanent error is permanent")
	}
	if IsPermanent(boom) || IsPermanent(nil) {
		t.Error("an ordinary error is not permanent")
	}
}
