package tasks

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/cygy/ginamite/worker/internal/workertest"
)

func TestLog(t *testing.T) {
	fields := []any{"count", 3}

	tests := []struct {
		name      string
		log       func()
		wantLevel slog.Level
		wantMsg   string
		wantExtra bool
		wantErr   string
	}{
		{"started", func() { LogStarted("t") }, slog.LevelInfo, "task started", false, ""},
		{"started with fields", func() { LogStarted("t", fields...) }, slog.LevelInfo, "task started", true, ""},
		{"done", func() { LogDone("t") }, slog.LevelInfo, "task done", false, ""},
		{"done with fields", func() { LogDone("t", fields...) }, slog.LevelInfo, "task done", true, ""},
		{"error", func() { LogError("t", errors.New("boom")) }, slog.LevelError, "task error", false, "boom"},
		{"error with fields", func() { LogError("t", errors.New("boom"), fields...) }, slog.LevelError, "task error", true, "boom"},
		{"error without error", func() { LogError("t", nil) }, slog.LevelError, "task error", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workertest.Hook.Reset()
			tt.log()

			entry := workertest.Hook.LastEntry()
			if entry == nil || entry.Level != tt.wantLevel || entry.Message != tt.wantMsg || entry.Data["task"] != "t" {
				t.Fatalf("entry = %+v", entry)
			}
			if _, ok := entry.Data["count"]; ok != tt.wantExtra {
				t.Errorf("fields = %v", entry.Data)
			}
			if got, _ := entry.Data["error"].(string); got != tt.wantErr {
				t.Errorf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}
