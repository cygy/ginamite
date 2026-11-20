package config

import (
	"testing"
	"time"
)

func TestGeoIPConfiguration_LocationCacheDuration(t *testing.T) {
	tests := map[time.Duration]time.Duration{
		0:                         DefaultGeoIPCacheDuration,
		-time.Hour:                DefaultGeoIPCacheDuration,
		time.Millisecond:          time.Second,
		6 * time.Hour:             6 * time.Hour,
		DefaultGeoIPCacheDuration: DefaultGeoIPCacheDuration,
	}
	for configured, want := range tests {
		if got := (GeoIPConfiguration{CacheDuration: configured}).LocationCacheDuration(); got != want {
			t.Errorf("LocationCacheDuration() with %s = %s, want %s", configured, got, want)
		}
	}
}
