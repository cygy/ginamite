package log_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/log/internal/reporter"
	"github.com/cygy/ginamite/common/log/logtest"
)

func TestDatabaseError_Caller(t *testing.T) {
	hook, restore := logtest.Replace(&log.Main, slog.LevelInfo)
	defer restore()

	log.DatabaseError(errors.New("boom"), "users")

	entry := hook.LastEntry()
	// The caller is the function which called DatabaseError: the test itself.
	if caller, _ := entry.Data["caller"].(string); !strings.HasSuffix(caller, "TestDatabaseError_Caller") {
		t.Errorf("caller = %v", entry.Data["caller"])
	}
	if source, _ := entry.Data["source"].(string); !strings.HasPrefix(source, "log/database_caller_test.go:") {
		t.Errorf("source = %v", entry.Data["source"])
	}
}

func TestDatabaseOperationError_Caller(t *testing.T) {
	hook, restore := logtest.Replace(&log.Main, slog.LevelInfo)
	defer restore()

	reporter.Report(context.Background(), errors.New("boom"))

	entry := hook.LastEntry()
	// The functions of the package of the reporter are skipped: the caller is the test, which called reporter.Report.
	if caller, _ := entry.Data["caller"].(string); !strings.HasSuffix(caller, "TestDatabaseOperationError_Caller") {
		t.Errorf("caller = %v", entry.Data["caller"])
	}
	if source, _ := entry.Data["source"].(string); !strings.HasPrefix(source, "log/database_caller_test.go:") {
		t.Errorf("source = %v", entry.Data["source"])
	}
}
