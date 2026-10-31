package account

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

var revoked = map[string]time.Time{"token": time.Unix(1700000000, 0)}

// deleter returns a function deleting a user, which records the users and revokes tokens. The user "bad" can not be
// deleted.
func deleter(deleted *[]string) func(context.Context, string) (map[string]time.Time, error) {
	return func(_ context.Context, userID string) (map[string]time.Time, error) {
		*deleted = append(*deleted, userID)
		if userID == "bad" {
			return nil, errors.New("boom")
		}
		if userID == "u1" {
			return revoked, nil
		}

		return nil, nil
	}
}

// checkDeletion checks the messages sent for the deletion of the users: the tokens of u1 are invalidated and
// the workers delete each user.
func checkDeletion(t *testing.T, q *workertest.Queue, userIDs ...string) {
	t.Helper()

	var got []string
	invalidated := false
	for {
		m, ok := q.Next()
		if !ok {
			break
		}

		switch m.Type {
		case queue.MessageCacheInvalidAuthToken:
			invalidated = true
		case queue.MessageTaskDeleteUser:
			got = append(got, m.Payload["user_id"].(string))
		default:
			t.Errorf("unexpected message %+v", m)
		}
	}

	// The producer is asynchronous: the order of the messages is not guaranteed.
	slices.Sort(got)
	if !slices.Equal(got, userIDs) {
		t.Errorf("deletions sent for %v, want %v", got, userIDs)
	}
	if want := slices.Contains(userIDs, "u1"); invalidated != want {
		t.Errorf("tokens invalidated = %v, want %v", invalidated, want)
	}
}

// checkCount checks the count logged by the task.
func checkCount(t *testing.T, want any) {
	t.Helper()

	for _, entry := range workertest.Hook.AllEntries() {
		if entry.Message == "task done" {
			if entry.Data["task"] != "task" || entry.Data["count"] != want {
				t.Errorf("entry = %+v, want count %v", entry, want)
			}

			return
		}
	}

	t.Error("the task is not logged as done")
}

func TestDeleteNeverUsed(t *testing.T) {
	q := workertest.CaptureQueue(t)
	workertest.Hook.Reset()

	var deleted []string
	var gotDays uint
	DeleteNeverUsed(context.Background(), q.Sender, "task", 10,
		func(_ context.Context, days uint) []string {
			gotDays = days
			return []string{"u1", "u2", "u3"}
		},
		func(_ context.Context, userID string) bool { return userID != "u2" },
		deleter(&deleted),
	)

	if gotDays != 10 {
		t.Errorf("interval = %d", gotDays)
	}
	if !slices.Equal(deleted, []string{"u1", "u3"}) {
		t.Errorf("deleted = %v, the user u2 is not empty", deleted)
	}
	checkDeletion(t, q, "u1", "u3")
	checkCount(t, 2)
}

func TestDeleteNeverUsed_Failure(t *testing.T) {
	q := workertest.CaptureQueue(t)
	workertest.Hook.Reset()

	var deleted []string
	DeleteNeverUsed(context.Background(), q.Sender, "task", 10,
		func(context.Context, uint) []string { return []string{"bad", "u1"} },
		func(context.Context, string) bool { return true },
		deleter(&deleted),
	)

	// The failure is logged, nothing is sent for the user, and the other users are deleted.
	checkDeletion(t, q, "u1")
	checkCount(t, 1)
	checkDeleteError(t, "bad")
}

// checkDeleteError checks that the failure to delete the user is logged.
func checkDeleteError(t *testing.T, userID string) {
	t.Helper()

	for _, entry := range workertest.Hook.AllEntries() {
		if entry.Message == "unable to delete the user" && entry.Data["user"] == userID && entry.Data["error"] == "boom" {
			return
		}
	}

	t.Errorf("the failure to delete %q is not logged", userID)
}

func TestDeleteInactive_Failure(t *testing.T) {
	q := workertest.CaptureQueue(t)
	workertest.Hook.Reset()

	var deleted []string
	DeleteInactive(context.Background(), q.Sender, "task", 48,
		func(context.Context, uint) []string { return []string{"bad", "u1"} },
		deleter(&deleted),
	)

	checkDeletion(t, q, "u1")
	checkCount(t, 1)
	checkDeleteError(t, "bad")
}

func TestDeleteInactive(t *testing.T) {
	q := workertest.CaptureQueue(t)
	workertest.Hook.Reset()

	var deleted []string
	var gotMonths uint
	DeleteInactive(context.Background(), q.Sender, "task", 48,
		func(_ context.Context, months uint) []string {
			gotMonths = months
			return []string{"u1", "u2"}
		},
		deleter(&deleted),
	)

	if gotMonths != 48 || !slices.Equal(deleted, []string{"u1", "u2"}) {
		t.Errorf("months = %d, deleted = %v", gotMonths, deleted)
	}
	checkDeletion(t, q, "u1", "u2")
	checkCount(t, 2)
}

func TestSanitize(t *testing.T) {
	workertest.Hook.Reset()

	Sanitize("task", func() uint { return 7 })

	checkCount(t, uint(7))
}
