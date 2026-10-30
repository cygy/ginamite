package recurring

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/worker/tasks"
)

// scheduledTasks are the scheduled tasks, by name.
// They are loaded before the tasks are started: the map is not protected.
var scheduledTasks = map[string]ScheduledTask{}

// wait returns a channel receiving the time after the duration. It is replaced by the tests.
var wait = time.After

// LoadScheduledTasks registers the scheduled tasks. The tasks without name are ignored.
func LoadScheduledTasks(tasks []ScheduledTask) {
	for _, task := range tasks {
		if len(task.Name) == 0 {
			continue
		}

		scheduledTasks[task.Name] = task

		log.WithField("name", task.Name).Info("scheduled task loaded")
	}
}

// SetFunc sets the function called by the scheduled task with that name. It does nothing if there is no such task.
func SetFunc(name string, f func(string)) {
	if task, ok := scheduledTasks[name]; ok {
		task.f = f
		scheduledTasks[name] = task
	}
}

// StartScheduledTasks starts the scheduled tasks which are enabled and have a function and an interval.
// Each task runs in the background, forever.
func StartScheduledTasks() {
	startScheduledTasks(context.Background())
}

// startScheduledTasks starts the scheduled tasks until the context is done.
func startScheduledTasks(ctx context.Context) {
	for name, task := range scheduledTasks {
		reason := ""
		switch {
		case !task.Enabled:
			reason = "not enabled"
		case task.f == nil:
			reason = "func not defined"
		case task.Interval == 0:
			reason = "interval not defined"
		}

		if reason != "" {
			log.WithField("name", name).WithField("reason", reason).Error("task not scheduled")
			continue
		}

		startIn := task.StartIn()
		interval := time.Duration(task.Interval) * time.Second

		go run(ctx, name, startIn, interval, task.f)

		log.WithField("name", name).WithField("start in", startIn).WithField("interval", interval).Info("task scheduled")
	}
}

// run calls f, in the background, after startIn then every interval, until the context is done.
func run(ctx context.Context, name string, startIn, interval time.Duration, f func(name string)) {
	delay := startIn

	for {
		select {
		case <-ctx.Done():
			return
		case <-wait(delay):
		}

		tasks.LogStarted(name, nil)
		go f(name)

		delay = interval
	}
}

// HasFunc reports whether the scheduled task with that name has a function.
func HasFunc(name string) bool {
	return scheduledTasks[name].f != nil
}
