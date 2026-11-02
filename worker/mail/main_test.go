package mail

import (
	"os"
	"testing"

	"github.com/cygy/ginamite/common/log"

	"github.com/cygy/ginamite/common/log/logtest"
)

// hook collects the entries of the logger.
var hook *logtest.Hook

func TestMain(m *testing.M) {
	log.Main, hook = logtest.New()

	os.Exit(m.Run())
}
