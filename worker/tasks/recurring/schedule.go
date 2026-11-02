package recurring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/worker/tasks"
)

// Func is the function of a scheduled task. It receives the context of the scheduler, done when the worker stops: a
// long task must return early once it is done.
type Func func(ctx context.Context, name string)

// Scheduler runs the scheduled tasks, each one in its own goroutine, until its context is done.
// The tasks are loaded before they are started: the scheduler is not safe for concurrent use. Create it with
// NewScheduler.
type Scheduler struct {
	tasks   map[string]ScheduledTask // The scheduled tasks, by name.
	running sync.WaitGroup           // The goroutines of the started tasks.
}

// NewScheduler returns a scheduler without task.
func NewScheduler() *Scheduler {
	return &Scheduler{tasks: map[string]ScheduledTask{}}
}

// wait returns a channel receiving the time after the duration. It is replaced by the tests.
var wait = time.After

// Load registers the tasks of the configuration. The tasks without name are ignored.
func (s *Scheduler) Load(tasks []config.RecurringTask) {
	for _, task := range tasks {
		if len(task.Name) == 0 {
			continue
		}

		s.tasks[task.Name] = ScheduledTask{RecurringTask: task}

		log.Info("scheduled task loaded", "name", task.Name)
	}
}

// SetFunc sets the function called by the scheduled task with that name. It does nothing if there is no such task.
func (s *Scheduler) SetFunc(name string, f Func) {
	if task, ok := s.tasks[name]; ok {
		task.f = f
		s.tasks[name] = task
	}
}

// Start starts the scheduled tasks which are enabled and have a function and an interval, until ctx is done. Each
// task runs in its own goroutine, one run at a time: a run longer than the interval delays the next one. Wait waits
// for the tasks once ctx is done.
func (s *Scheduler) Start(ctx context.Context) {
	for name, task := range s.tasks {
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
			log.Error("task not scheduled",
				"name", name,
				"reason", reason,
			)
			continue
		}

		startIn := task.StartIn()
		interval := time.Duration(task.Interval) * time.Second //nolint:gosec // an interval in seconds is far below the overflow

		s.running.Go(func() { run(ctx, name, startIn, interval, task.f) })

		log.Info("task scheduled",
			"name", name,
			"start in", startIn,
			"interval", interval,
		)
	}
}

// Wait waits for the goroutines of the tasks, which return once the context given to Start is done and their running
// run, if any, returned.
func (s *Scheduler) Wait() {
	s.running.Wait()
}

// run calls f after startIn then every interval, from the start of a run to the start of the next one, until the
// context is done. A run longer than the interval is followed by the next one at once.
func run(ctx context.Context, name string, startIn, interval time.Duration, f Func) {
	delay := startIn

	for {
		select {
		case <-ctx.Done():
			return
		case <-wait(delay):
		}

		started := time.Now()
		execute(ctx, name, f)

		delay = max(interval-time.Since(started), 0)
	}
}

// execute runs f, and logs its panic instead of stopping the worker.
func execute(ctx context.Context, name string, f Func) {
	defer func() {
		if r := recover(); r != nil {
			tasks.LogError(name, fmt.Errorf("the task panicked: %v", r))
		}
	}()

	tasks.LogStarted(name)
	f(ctx, name)
}

// HasFunc reports whether the scheduled task with that name has a function.
func (s *Scheduler) HasFunc(name string) bool {
	return s.tasks[name].f != nil
}
