package recurring

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cygy/ginamite/worker/internal/workertest"
)

// resetTasks empties the scheduled tasks, for the duration of the test.
func resetTasks(t *testing.T) {
	t.Helper()

	old := scheduledTasks
	t.Cleanup(func() { scheduledTasks = old })
	scheduledTasks = map[string]ScheduledTask{}
}

func TestLoadScheduledTasks(t *testing.T) {
	resetTasks(t)

	LoadScheduledTasks([]ScheduledTask{{Name: "a", Interval: 10}, {Name: ""}, {Name: "b", Enabled: true}})

	if len(scheduledTasks) != 2 || scheduledTasks["a"].Interval != 10 || !scheduledTasks["b"].Enabled {
		t.Errorf("scheduledTasks = %+v", scheduledTasks)
	}
}

func TestSetFunc(t *testing.T) {
	resetTasks(t)
	LoadScheduledTasks([]ScheduledTask{{Name: "a"}})

	SetFunc("a", func(string) {})
	SetFunc("unknown", func(string) {})

	if scheduledTasks["a"].f == nil {
		t.Error("the function of the task is not set")
	}
	if _, ok := scheduledTasks["unknown"]; ok {
		t.Error("a task must not be created by SetFunc")
	}
}

func TestStartScheduledTasks(t *testing.T) {
	resetTasks(t)

	oldWait := wait
	t.Cleanup(func() { wait = oldWait })

	var mu sync.Mutex
	var delays []time.Duration
	ticks := make(chan time.Time)
	wait = func(d time.Duration) <-chan time.Time {
		mu.Lock()
		delays = append(delays, d)
		mu.Unlock()

		return ticks
	}

	runs := make(chan string, 10)
	LoadScheduledTasks([]ScheduledTask{
		{Name: "ok", Interval: 5, StartAt: "bad", Enabled: true},
		{Name: "disabled", Interval: 5},
		{Name: "no func", Interval: 5, Enabled: true},
		{Name: "no interval", Enabled: true},
	})
	SetFunc("ok", func(name string) { runs <- name })
	SetFunc("disabled", func(string) { t.Error("a disabled task must not run") })
	SetFunc("no interval", func(string) { t.Error("a task without interval must not run") })

	workertest.Hook.Reset()
	ctx, cancel := context.WithCancel(context.Background())
	startScheduledTasks(ctx)

	for range 3 {
		ticks <- time.Now()

		select {
		case name := <-runs:
			if name != "ok" {
				t.Errorf("task %q ran", name)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("the task did not run")
		}
	}
	cancel()

	// The next wait is interrupted by the cancellation.
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(delays) < 3 || delays[0] != defaultStartIn || delays[1] != 5*time.Second || delays[2] != 5*time.Second {
		t.Errorf("delays = %v", delays)
	}

	reasons := map[string]string{}
	for _, entry := range workertest.Hook.AllEntries() {
		if entry.Message == "task not scheduled" {
			reasons[entry.Data["name"].(string)] = entry.Data["reason"].(string)
		}
	}
	want := map[string]string{"disabled": "not enabled", "no func": "func not defined", "no interval": "interval not defined"}
	if len(reasons) != len(want) {
		t.Fatalf("reasons = %v, want %v", reasons, want)
	}
	for name, reason := range want {
		if reasons[name] != reason {
			t.Errorf("reason of %q = %q, want %q", name, reasons[name], reason)
		}
	}
}

func TestStartScheduledTasks_Default(t *testing.T) {
	resetTasks(t)

	// No task: nothing is started, and the default context is used.
	StartScheduledTasks()
}

func TestHasFunc(t *testing.T) {
	resetTasks(t)
	LoadScheduledTasks([]ScheduledTask{{Name: "a"}, {Name: "b"}})
	SetFunc("a", func(string) {})

	if !HasFunc("a") || HasFunc("b") || HasFunc("unknown") {
		t.Errorf("HasFunc: a = %v, b = %v, unknown = %v", HasFunc("a"), HasFunc("b"), HasFunc("unknown"))
	}
}
