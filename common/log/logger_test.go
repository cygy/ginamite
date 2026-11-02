package log

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestInitialize(t *testing.T) {
	defer func(old *slog.Logger) { Main = old }(Main)

	tests := []struct {
		name       string
		production bool
		wantJSON   bool
	}{
		{"development", false, false},
		{"production", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "log")
			if err != nil {
				t.Fatal(err)
			}
			defer func(old *os.File) { os.Stdout = old }(os.Stdout)
			os.Stdout = output

			Initialize("srv", tt.production)
			Info("m", "k", "v")

			written, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}

			var record map[string]any
			isJSON := json.Unmarshal(written, &record) == nil
			if isJSON != tt.wantJSON {
				t.Errorf("JSON output = %v, want %v: %s", isJSON, tt.wantJSON, written)
			}
			if !strings.Contains(string(written), "srv") || !strings.Contains(string(written), "k") {
				t.Errorf("the server name and the attributes must be logged: %s", written)
			}
		})
	}
}

func TestShorthands(t *testing.T) {
	hook := useHook(t)

	tests := []struct {
		name  string
		log   func()
		level slog.Level
	}{
		{"Debug", func() { Debug("m") }, slog.LevelDebug},
		{"Info", func() { Info("m") }, slog.LevelInfo},
		{"Warn", func() { Warn("m") }, slog.LevelWarn},
		{"Error", func() { Error("m") }, slog.LevelError},
		{"With", func() { With("k", "v").Info("m") }, slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()
			tt.log()

			entry := hook.LastEntry()
			if entry == nil || entry.Level != tt.level || entry.Message != "m" {
				t.Fatalf("entry = %+v, want level %v", entry, tt.level)
			}
			if entry.Data["server"] != "test" {
				t.Errorf("the main logger fields are missing: %v", entry.Data)
			}
		})
	}
}
