package recurring

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cygy/ginamite/worker/internal/workertest"
)

func TestScheduler_Load(t *testing.T) {
	s := NewScheduler()

	s.Load([]ScheduledTask{{Name: "a", Interval: 10}, {Name: ""}, {Name: "b", Enabled: true}})

	if len(s.tasks) != 2 || s.tasks["a"].Interval != 10 || !s.tasks["b"].Enabled {
		t.Errorf("s.tasks = %+v", s.tasks)
	}
}

func TestScheduler_SetFunc(t *testing.T) {
	s := NewScheduler()
	s.Load([]ScheduledTask{{Name: "a"}})

	s.SetFunc("a", func(string) {})
	s.SetFunc("unknown", func(string) {})

	if s.tasks["a"].f == nil {
		t.Error("the function of the task is not set")
	}
	if _, ok := s.tasks["unknown"]; ok {
		t.Error("a task must not be created by SetFunc")
	}
}

func TestScheduler_Start(t *testing.T) {
	s := NewScheduler()

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
	s.Load([]ScheduledTask{
		{Name: "ok", Interval: 5, StartAt: "bad", Enabled: true},
		{Name: "disabled", Interval: 5},
		{Name: "no func", Interval: 5, Enabled: true},
		{Name: "no interval", Enabled: true},
	})
	s.SetFunc("ok", func(name string) { runs <- name })
	s.SetFunc("disabled", func(string) { t.Error("a disabled task must not run") })
	s.SetFunc("no interval", func(string) { t.Error("a task without interval must not run") })

	workertest.Hook.Reset()
	ctx, cancel := context.WithCancel(context.Background())
	s.start(ctx)

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

func TestScheduler_Start_Default(t *testing.T) {
	s := NewScheduler()

	// No task: nothing is started, and the default context is used.
	s.Start()
}

func TestScheduler_HasFunc(t *testing.T) {
	s := NewScheduler()
	s.Load([]ScheduledTask{{Name: "a"}, {Name: "b"}})
	s.SetFunc("a", func(string) {})

	if !s.HasFunc("a") || s.HasFunc("b") || s.HasFunc("unknown") {
		t.Errorf("HasFunc: a = %v, b = %v, unknown = %v", s.HasFunc("a"), s.HasFunc("b"), s.HasFunc("unknown"))
	}
}
