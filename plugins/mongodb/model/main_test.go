package model

import (
	"log/slog"
	"os"
	"testing"

	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Main = slog.New(slog.DiscardHandler)
	os.Exit(m.Run())
}
