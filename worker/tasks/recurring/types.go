package recurring

import (
	"strconv"
	"strings"
	"time"
)

// ScheduledTask : a scheduled task that start the first time at 'StartAt' and run every 'Interval' seconds after that.
type ScheduledTask struct {
	Name     string `yaml:"name"`
	Interval uint   `yaml:"interval"`
	StartAt  string `yaml:"start_at"` // must be an hour i.e. "14:30" or "4:15am"
	Enabled  bool   `yaml:"enabled"`

	// This is the func called every 'Interval' seconds.
	f func(taskName string)
}

// StartIn : returns the duration to wait of the first launch of the function. By default it returns 60 seconds.
func (task *ScheduledTask) StartIn() time.Duration {
	t := strings.ToLower(task.StartAt)
	isAM := strings.HasSuffix(t, "am")
	isPM := strings.HasSuffix(t, "pm")

	t = strings.ReplaceAll(t, "am", "")
	t = strings.ReplaceAll(t, "pm", "")

	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return 60 * time.Second
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])

	if isAM && hours == 12 {
		hours = 0
	}
	if isPM && hours < 12 {
		hours += 12
	}

	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hours, minutes, 0, 0, time.UTC)

	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	return next.Sub(now)
}
