package recurring

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

func TestScheduler_Load(t *testing.T) {
	s := NewScheduler()

	s.Load([]config.RecurringTask{{Name: "a", Interval: 10}, {Name: ""}, {Name: "b", Enabled: true}})

	if len(s.tasks) != 2 || s.tasks["a"].Interval != 10 || !s.tasks["b"].Enabled {
		t.Errorf("s.tasks = %+v", s.tasks)
	}
}

func TestScheduler_SetFunc(t *testing.T) {
	s := NewScheduler()
	s.Load([]config.RecurringTask{{Name: "a"}})

	s.SetFunc("a", func(context.Context, string) {})
	s.SetFunc("unknown", func(context.Context, string) {})

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
	s.Load([]config.RecurringTask{
		{Name: "ok", Interval: 5, StartAt: "bad", Enabled: true},
		{Name: "disabled", Interval: 5},
		{Name: "no func", Interval: 5, Enabled: true},
		{Name: "no interval", Enabled: true},
	})
	s.SetFunc("ok", func(_ context.Context, name string) { runs <- name })
	s.SetFunc("disabled", func(context.Context, string) { t.Error("a disabled task must not run") })
	s.SetFunc("no interval", func(context.Context, string) { t.Error("a task without interval must not run") })

	workertest.Hook.Reset()
	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

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
	// The next wait is interrupted by the cancellation, and Wait returns.
	cancel()
	s.Wait()

	mu.Lock()
	defer mu.Unlock()
	// The interval runs from the start of a run: the time of the run is taken from the next delay.
	if len(delays) < 3 || delays[0] != defaultStartIn || !nearInterval(delays[1]) || !nearInterval(delays[2]) {
		t.Errorf("delays = %v", delays)
	}

	reasons := map[string]string{}
	levels := map[string]slog.Level{}
	for _, entry := range workertest.Hook.AllEntries() {
		if entry.Message == "task not scheduled" {
			name := entry.Data["name"].(string)
			reasons[name] = entry.Data["reason"].(string)
			levels[name] = entry.Level
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
	// A disabled task is expected, the others are errors.
	wantLevels := map[string]slog.Level{"disabled": slog.LevelInfo, "no func": slog.LevelError, "no interval": slog.LevelError}
	for name, level := range wantLevels {
		if levels[name] != level {
			t.Errorf("level of %q = %v, want %v", name, levels[name], level)
		}
	}
}

// nearInterval reports whether d is the interval of 5 seconds minus the short time of a run.
func nearInterval(d time.Duration) bool {
	return d <= 5*time.Second && d > 4*time.Second
}

func TestScheduler_NoTask(t *testing.T) {
	s := NewScheduler()

	// No task: nothing is started, and Wait returns at once.
	s.Start(t.Context())
	s.Wait()
}

func TestScheduler_PanicAndLongRun(t *testing.T) {
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

	runs := make(chan int, 10)
	count := 0
	s := NewScheduler()
	s.Load([]config.RecurringTask{{Name: "task", Interval: 1, Enabled: true}})
	s.SetFunc("task", func(context.Context, string) {
		count++
		runs <- count
		if count == 1 {
			panic("boom")
		}
		time.Sleep(1100 * time.Millisecond) // Longer than the interval.
	})

	workertest.Hook.Reset()
	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

	// The panic of the first run is logged, and the task runs again.
	for want := 1; want <= 2; want++ {
		ticks <- time.Now()
		if got := <-runs; got != want {
			t.Fatalf("run = %d, want %d", got, want)
		}
	}
	cancel()
	s.Wait()

	mu.Lock()
	defer mu.Unlock()
	// After the long run the next one is due at once.
	if len(delays) != 3 || delays[2] != 0 {
		t.Errorf("delays = %v, want no delay after a run longer than the interval", delays)
	}

	panicked := false
	for _, entry := range workertest.Hook.AllEntries() {
		if e, ok := entry.Data["error"].(string); ok && e == "the task panicked: boom" {
			panicked = true
		}
	}
	if !panicked {
		t.Error("the panic of the task must be logged")
	}
}

func TestScheduler_WaitForTheRunningRun(t *testing.T) {
	oldWait := wait
	t.Cleanup(func() { wait = oldWait })

	ticks := make(chan time.Time, 1)
	wait = func(time.Duration) <-chan time.Time { return ticks }

	started := make(chan struct{})
	finished := false
	s := NewScheduler()
	s.Load([]config.RecurringTask{{Name: "task", Interval: 60, Enabled: true}})
	s.SetFunc("task", func(ctx context.Context, _ string) {
		close(started)
		<-ctx.Done() // A task returns once the worker stops.
		finished = true
	})

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)
	ticks <- time.Now()
	<-started

	cancel()
	s.Wait()

	if !finished {
		t.Error("Wait must wait for the running run")
	}
}

func TestScheduler_HasFunc(t *testing.T) {
	s := NewScheduler()
	s.Load([]config.RecurringTask{{Name: "a"}, {Name: "b"}})
	s.SetFunc("a", func(context.Context, string) {})

	if !s.HasFunc("a") || s.HasFunc("b") || s.HasFunc("unknown") {
		t.Errorf("HasFunc: a = %v, b = %v, unknown = %v", s.HasFunc("a"), s.HasFunc("b"), s.HasFunc("unknown"))
	}
}
