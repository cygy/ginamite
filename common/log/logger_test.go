package log

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInitialize(t *testing.T) {
	defer func(old *logrus.Entry) { Main = old }(Main)

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
			Initialize("srv", tt.production)

			if Main.Data["server"] != "srv" {
				t.Errorf("server field = %v", Main.Data["server"])
			}
			_, isJSON := Main.Logger.Formatter.(*logrus.JSONFormatter)
			if isJSON != tt.wantJSON {
				t.Errorf("JSON formatter = %v, want %v", isJSON, tt.wantJSON)
			}
		})
	}
}

func TestShorthands(t *testing.T) {
	hook := useHook(t)

	tests := []struct {
		name  string
		log   func()
		level logrus.Level
	}{
		{"Debug", func() { Debug("m") }, logrus.DebugLevel},
		{"Info", func() { Info("m") }, logrus.InfoLevel},
		{"Print", func() { Print("m") }, logrus.InfoLevel},
		{"Warn", func() { Warn("m") }, logrus.WarnLevel},
		{"Warning", func() { Warning("m") }, logrus.WarnLevel},
		{"Error", func() { Error("m") }, logrus.ErrorLevel},
		{"Fatal", func() { Fatal("m") }, logrus.FatalLevel},
		{"WithField", func() { WithField("k", "v").Info("m") }, logrus.InfoLevel},
		{"WithFields", func() { WithFields(logrus.Fields{"k": "v"}).Info("m") }, logrus.InfoLevel},
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

func TestPanic(t *testing.T) {
	hook := useHook(t)

	defer func() {
		if recover() == nil {
			t.Error("Panic() must panic")
		}
		if entry := hook.LastEntry(); entry == nil || entry.Level != logrus.PanicLevel {
			t.Errorf("entry = %+v, want a panic entry", entry)
		}
	}()

	Panic("m")
}
