package recurring

import (
	"time"

	"github.com/cygy/ginamite/worker/tasks"

	"github.com/cygy/ginamite/common/log"
)

// LoadScheduledTasks : register scheduled tasks.
func LoadScheduledTasks(tasks []ScheduledTask) {
	for _, task := range tasks {
		if len(task.Name) == 0 {
			continue
		}

		scheduledTasks[task.Name] = task

		log.WithField("name", task.Name).Info("scheduled task loaded")
	}
}

// SetFunc : saves a func to the scheduled task with that name.
func SetFunc(name string, f func(string)) {
	if task, ok := scheduledTasks[name]; ok {
		task.f = f
		scheduledTasks[name] = task
	}
}

// StartScheduledTasks : start the scheduled tasks.
func StartScheduledTasks() {
	run := func(name string, startIn, interval time.Duration, f func(name string)) {
		<-time.After(startIn)

		for {
			tasks.LogStarted(name, nil)
			go f(name)
			<-time.After(interval)
		}
	}

	for name, task := range scheduledTasks {
		if !task.Enabled {
			log.WithField("name", name).WithField("reason", "not enabled").Error("task not scheduled")
			continue
		}
		if task.f == nil {
			log.WithField("name", name).WithField("reason", "func not defined").Error("task not scheduled")
			continue
		}

		startIn := task.StartIn()
		interval := time.Duration(task.Interval) * time.Second

		go run(name, startIn, interval, task.f)

		log.WithField("name", name).WithField("start in", startIn).WithField("interval", interval).Info("task scheduled")
	}
}
