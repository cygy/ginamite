package logtest

import (
	"log/slog"
	"testing"
)

func TestNew(t *testing.T) {
	logger, hook := New()

	logger.Debug("hidden")
	logger.With("a", 1).WithGroup("g").Warn("m", "uint", uint(2), slog.Group("n", "k", "v"), "s", "x")

	entries := hook.AllEntries()
	if len(entries) != 1 {
		t.Fatalf("entries = %+v, want only the warning", entries)
	}

	entry := hook.LastEntry()
	if entry.Level != slog.LevelWarn || entry.Message != "m" || entry.Data["server"] != "test" || entry.Data["a"] != 1 {
		t.Errorf("entry = %+v", entry)
	}

	group, _ := entry.Data["g"].(map[string]any)
	nested, _ := group["n"].(map[string]any)
	if group["uint"] != uint(2) || group["s"] != "x" || nested["k"] != "v" {
		t.Errorf("group = %+v", group)
	}

	hook.Reset()
	if hook.LastEntry() != nil || len(hook.AllEntries()) != 0 {
		t.Error("Reset must forget the entries")
	}
}

func TestNewWithLevel(t *testing.T) {
	logger, hook := NewWithLevel(slog.LevelDebug)

	logger.Debug("shown")

	if entry := hook.LastEntry(); entry == nil || entry.Message != "shown" {
		t.Errorf("entry = %+v", entry)
	}
}

func TestReplace(t *testing.T) {
	logger, _ := New()
	original := logger

	hook, restore := Replace(&logger, slog.LevelInfo)
	logger.Info("m")

	if logger == original || hook.LastEntry() == nil {
		t.Error("the logger must be replaced by one collecting its entries")
	}

	restore()
	if logger != original {
		t.Error("restore must put the previous logger back")
	}
}
