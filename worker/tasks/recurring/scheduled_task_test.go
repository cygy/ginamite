package recurring

import (
	"testing"
	"time"
)

func TestScheduledTask_StartIn(t *testing.T) {
	// The current time is 2026-10-30 10:00:00 UTC, whatever the local time zone is.
	oldNow := now
	t.Cleanup(func() { now = oldNow })
	now = func() time.Time { return time.Date(2026, 10, 30, 11, 0, 0, 0, time.FixedZone("CET", 3600)) }

	tests := []struct {
		startAt string
		want    time.Duration
	}{
		{"", defaultStartIn},
		{"14:30", 4*time.Hour + 30*time.Minute},
		{"9:00", 23 * time.Hour},
		{"10:00", 0},
		{"4:15am", 18*time.Hour + 15*time.Minute},
		{"4:15PM", 6*time.Hour + 15*time.Minute},
		{"12:30am", 14*time.Hour + 30*time.Minute},
		{"12:30pm", 2*time.Hour + 30*time.Minute},
		{" 14:30 ", 4*time.Hour + 30*time.Minute},
		{"1430", defaultStartIn},
		{"14:30:00", defaultStartIn},
		{"ab:cd", defaultStartIn},
		{"14:xx", defaultStartIn},
		{"24:00", defaultStartIn},
		{"-1:00", defaultStartIn},
		{"10:60", defaultStartIn},
		{"10:-5", defaultStartIn},
	}

	for _, tt := range tests {
		t.Run(tt.startAt, func(t *testing.T) {
			task := ScheduledTask{StartAt: tt.startAt}

			if got := task.StartIn(); got != tt.want {
				t.Errorf("StartIn() = %v, want %v", got, tt.want)
			}
		})
	}
}
