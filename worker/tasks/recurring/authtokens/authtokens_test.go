package authtokens

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/worker/internal/workertest"
)

func TestDeleteExpired(t *testing.T) {
	workertest.Hook.Reset()

	DeleteExpired(context.Background(), "delete_expired_tokens", func(context.Context) int64 { return 4 })

	entry := workertest.Hook.LastEntry()
	if entry == nil || entry.Message != "task done" || entry.Data["task"] != "delete_expired_tokens" || entry.Data["count"] != int64(4) {
		t.Errorf("entry = %+v", entry)
	}
}
