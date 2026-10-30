package model

import (
	"io"
	"os"
	"testing"

	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	log.Main.Logger.SetOutput(io.Discard)
	os.Exit(m.Run())
}
