package authtokens

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cygy/ginamite/worker/internal/workertest"
)

func TestDeleteExpired(t *testing.T) {
	workertest.Hook.Reset()

	DeleteExpired(context.Background(), "delete_expired_tokens", func(context.Context) (int64, error) { return 4, nil })

	entry := workertest.Hook.LastEntry()
	if entry == nil || entry.Message != "task done" || entry.Data["task"] != "delete_expired_tokens" || entry.Data["count"] != 4 {
		t.Errorf("entry = %+v", entry)
	}
}

func TestDeleteExpired_Failure(t *testing.T) {
	workertest.Hook.Reset()

	DeleteExpired(context.Background(), "delete_expired_tokens", func(context.Context) (int64, error) { return 0, errors.New("boom") })

	entry := workertest.Hook.LastEntry()
	if entry == nil || entry.Message != "task error" || !strings.Contains(fmt.Sprint(entry.Data["error"]), "boom") {
		t.Errorf("entry = %+v", entry)
	}
}
