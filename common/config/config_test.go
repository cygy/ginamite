package config

import (
	"testing"

	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"
)

func TestInitialize(t *testing.T) {
	defer func(old *Configuration) { Main = old }(Main)

	if err := Initialize(writeFile(t, "default_locale: fr-fr\n"), environment.Test, version.Alpha); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if Main == nil || Main.DefaultLocale != "fr-fr" || !Main.IsTestEnvironment() || !Main.IsAlphaVersion() {
		t.Errorf("Main = %+v", Main)
	}
}
