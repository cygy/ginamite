package tasks

import (
	"errors"
	"testing"

	"github.com/cygy/ginamite/worker/internal/workertest"

	"github.com/sirupsen/logrus"
)

func TestLog(t *testing.T) {
	fields := &logrus.Fields{"count": 3}

	tests := []struct {
		name      string
		log       func()
		wantLevel logrus.Level
		wantMsg   string
		wantExtra bool
		wantErr   string
	}{
		{"started", func() { LogStarted("t", nil) }, logrus.InfoLevel, "task started", false, ""},
		{"started with fields", func() { LogStarted("t", fields) }, logrus.InfoLevel, "task started", true, ""},
		{"done", func() { LogDone("t", nil) }, logrus.InfoLevel, "task done", false, ""},
		{"done with fields", func() { LogDone("t", fields) }, logrus.InfoLevel, "task done", true, ""},
		{"error", func() { LogError("t", errors.New("boom"), nil) }, logrus.ErrorLevel, "task error", false, "boom"},
		{"error with fields", func() { LogError("t", errors.New("boom"), fields) }, logrus.ErrorLevel, "task error", true, "boom"},
		{"error without error", func() { LogError("t", nil, nil) }, logrus.ErrorLevel, "task error", false, ""},
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
