package html

import (
	"sync"
	"testing"
	"time"
)

func TestTimeInTimezone(t *testing.T) {
	utc := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		timezone string
		wantHour int
		wantErr  bool
	}{
		{"summer time", "Europe/Paris", 14, false},
		{"utc", "UTC", 12, false},
		{"empty is utc", "", 12, false},
		{"unknown", "Mars/Olympus", 12, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := timeInTimezone(utc, tt.timezone)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, want error: %v", err, tt.wantErr)
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("hour = %d, want %d", got.Hour(), tt.wantHour)
			}
			if !got.Equal(utc) {
				t.Errorf("%v is not the same instant as %v", got, utc)
			}
		})
	}
}

func TestTimeInTimezone_Concurrent(t *testing.T) {
	zones := []string{"Europe/Paris", "Asia/Tokyo", "America/New_York", "Africa/Lagos"}

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := timeInTimezone(time.Now(), zones[i%len(zones)]); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
