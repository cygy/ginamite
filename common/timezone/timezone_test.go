package timezone

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

const sample = `# comment
FR	+4852+00220	Europe/Paris
BE,LU,NL	+5050+00420	Europe/Brussels
FR,CH	+4622+00608	Europe/Zurich	with a comment
malformed line
CH	+4622+00608	Europe/Brussels
`

func writeSample(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "zones.tab")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func load(t *testing.T) *Configuration {
	t.Helper()

	conf := NewConfiguration()
	if err := conf.Initialize(writeSample(t, sample)); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	return conf
}

func TestNewConfiguration(t *testing.T) {
	conf := NewConfiguration()

	if conf.TimezonesByCountries == nil || conf.CountriesByTimezones == nil || conf.Countries == nil || conf.Timezones == nil {
		t.Errorf("the collections must be initialized: %+v", conf)
	}
}

func TestConfiguration_Initialize(t *testing.T) {
	conf := load(t)

	if want := []string{"BE", "CH", "FR", "LU", "NL"}; !reflect.DeepEqual(conf.Countries, want) {
		t.Errorf("Countries = %v, want %v", conf.Countries, want)
	}
	if want := []string{"Europe/Brussels", "Europe/Paris", "Europe/Zurich"}; !reflect.DeepEqual(conf.Timezones, want) {
		t.Errorf("Timezones = %v, want %v", conf.Timezones, want)
	}
	if want := []string{"Europe/Paris", "Europe/Zurich"}; !reflect.DeepEqual(conf.TimezonesByCountries["FR"], want) {
		t.Errorf("FR = %v, want %v", conf.TimezonesByCountries["FR"], want)
	}
	if want := []string{"BE", "CH", "LU", "NL"}; !reflect.DeepEqual(conf.CountriesByTimezones["Europe/Brussels"], want) {
		t.Errorf("Europe/Brussels = %v, want %v", conf.CountriesByTimezones["Europe/Brussels"], want)
	}
}

func TestConfiguration_Initialize_Errors(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Initialize() must panic after logging the error")
		}
	}()

	_ = NewConfiguration().Initialize(filepath.Join(t.TempDir(), "missing.tab"))
}

func TestLoad(t *testing.T) {
	conf, err := Load(writeSample(t, sample))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !conf.SupportsTimezone("Europe/Paris") {
		t.Errorf("conf = %+v", conf)
	}
}

func TestConfiguration_Initialize_RealFile(t *testing.T) {
	conf := NewConfiguration()
	if err := conf.Initialize("../../data/zone1970.tab"); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if !conf.SupportsTimezone("Europe/Paris") || len(conf.GetTimezones("fr")) == 0 {
		t.Error("the shipped timezone file must be loadable")
	}
}

func TestConfiguration_GetTimezones(t *testing.T) {
	conf := load(t)

	tests := []struct {
		country string
		want    []string
	}{
		{"FR", []string{"Europe/Paris", "Europe/Zurich"}},
		{"fr", []string{"Europe/Paris", "Europe/Zurich"}},
		{"XX", []string{}},
		{"", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			if got := conf.GetTimezones(tt.country); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTimezones(%q) = %v, want %v", tt.country, got, tt.want)
			}
		})
	}
}

func TestConfiguration_GetCountries(t *testing.T) {
	conf := load(t)

	tests := []struct {
		timezone string
		want     []string
	}{
		{"Europe/Paris", []string{"FR"}},
		{"Europe/Zurich", []string{"CH", "FR"}},
		{"Nowhere/Land", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.timezone, func(t *testing.T) {
			if got := conf.GetCountries(tt.timezone); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCountries(%q) = %v, want %v", tt.timezone, got, tt.want)
			}
		})
	}
}

func TestConfiguration_SupportsTimezone(t *testing.T) {
	conf := load(t)

	if !conf.SupportsTimezone("Europe/Paris") || conf.SupportsTimezone("europe/paris") || conf.SupportsTimezone("") {
		t.Error("SupportsTimezone() must be exact")
	}
}

func TestHasValue(t *testing.T) {
	if !hasValue([]string{"a", "b"}, "b") || hasValue([]string{"a"}, "c") || hasValue(nil, "a") {
		t.Error("hasValue() is wrong")
	}
}
