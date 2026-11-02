package recurring

import (
	"strconv"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/config"
)

// defaultStartIn is the delay before the first run of a task without a valid start time.
const defaultStartIn = 60 * time.Second

// now returns the current time. It is replaced by the tests.
var now = time.Now

// ScheduledTask is a task of the configuration, running the first time at StartAt, then every Interval seconds.
type ScheduledTask struct {
	config.RecurringTask

	// f is called every Interval seconds, with the name of the task.
	f func(taskName string)
}

// StartIn returns the duration to wait before the first run of the task: the time until the next StartAt, today or
// tomorrow. It is 60 seconds if StartAt is not a valid hour.
func (task *ScheduledTask) StartIn() time.Duration {
	t := strings.ToLower(strings.TrimSpace(task.StartAt))
	isAM := strings.HasSuffix(t, "am")
	isPM := strings.HasSuffix(t, "pm")

	t = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(t, "am"), "pm"))

	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return defaultStartIn
	}

	hours, errHours := strconv.Atoi(parts[0])
	minutes, errMinutes := strconv.Atoi(parts[1])
	if errHours != nil || errMinutes != nil || hours < 0 || minutes < 0 || minutes > 59 {
		return defaultStartIn
	}

	if isAM && hours == 12 {
		hours = 0
	}
	if isPM && hours < 12 {
		hours += 12
	}
	if hours > 23 {
		return defaultStartIn
	}

	current := now().UTC()
	next := time.Date(current.Year(), current.Month(), current.Day(), hours, minutes, 0, 0, time.UTC)

	if next.Before(current) {
		next = next.Add(24 * time.Hour)
	}

	return next.Sub(current)
}
