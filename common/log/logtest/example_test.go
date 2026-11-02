package logtest_test

import (
	"fmt"
	"log/slog"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/log/logtest"
)

// A test replaces the main logger to check the records of the code under test.
func ExampleReplace() {
	hook, restore := logtest.Replace(&log.Main, slog.LevelInfo)
	defer restore()

	log.Info("user created", "id", 42)

	entry := hook.LastEntry()
	fmt.Println(entry.Level, entry.Message, entry.Data["id"])
	// Output:
	// INFO user created 42
}
