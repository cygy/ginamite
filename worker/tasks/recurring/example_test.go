package recurring_test

import (
	"context"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/worker/tasks/recurring"
)

// The tasks of the configuration run until the context is done; Wait waits for the running ones.
func ExampleScheduler() {
	scheduler := recurring.NewScheduler()
	scheduler.Load([]config.RecurringTask{{Name: "report", Interval: 3600, StartAt: "02:00", Enabled: true}})
	scheduler.SetFunc("report", func(ctx context.Context, name string) {
		// Send the report, and return early once ctx is done.
	})

	ctx, cancel := context.WithCancel(context.Background())
	scheduler.Start(ctx)

	// When the worker stops:
	cancel()
	scheduler.Wait()
}
