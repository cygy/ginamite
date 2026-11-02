package authentication

import (
	"os"
	"testing"

	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}
