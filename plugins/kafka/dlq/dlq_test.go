package dlq

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

// fakeBackend keeps dead letters in memory.
type fakeBackend struct {
	messages map[int32][]*sarama.ConsumerMessage // By partition, the offset of a message is its index.
	next     map[int32]int64                     // The saved position of each partition.

	published  []*sarama.ProducerMessage
	publishErr error // Returned for the message with this value.
	failValue  string

	commits    []string
	partitions error
	position   error
	commitErr  error
}

func newBackend(counts ...int) *fakeBackend {
	b := &fakeBackend{messages: map[int32][]*sarama.ConsumerMessage{}, next: map[int32]int64{}}
	for partition, count := range counts {
		for offset := range count {
			b.messages[int32(partition)] = append(b.messages[int32(partition)], &sarama.ConsumerMessage{
				Topic:     "app_mail_dlq",
				Partition: int32(partition),
				Offset:    int64(offset),
				Timestamp: time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC),
				Key:       []byte("key"),
				Value:     fmt.Appendf(nil, "body-%d-%d", partition, offset),
				Headers: []*sarama.RecordHeader{
					{Key: []byte("error"), Value: []byte("smtp down")},
					{Key: []byte("attempts"), Value: []byte("3")},
					{Key: []byte("group"), Value: []byte("app_mail")},
					{Key: []byte("topic"), Value: []byte("app_mail")},
					{Key: []byte("partition"), Value: []byte("1")},
					{Key: []byte("offset"), Value: []byte("340")},
				},
			})
		}
	}

	return b
}

func (b *fakeBackend) Partitions(string) ([]int32, error) {
	if b.partitions != nil {
		return nil, b.partitions
	}

	var partitions []int32
	for partition := range b.messages {
		partitions = append(partitions, partition)
	}
	slices.Sort(partitions)

	return partitions, nil
}

func (b *fakeBackend) Position(_ string, partition int32) (int64, int64, error) {
	if b.position != nil {
		return 0, 0, b.position
	}

	return b.next[partition], int64(len(b.messages[partition])), nil
}

func (b *fakeBackend) Read(_ context.Context, _ string, partition int32, from, to int64, fn func(*sarama.ConsumerMessage) error) error {
	for _, message := range b.messages[partition][from:to] {
		if err := fn(message); err != nil {
			return err
		}
	}

	return nil
}

func (b *fakeBackend) Commit(_ string, partition int32, next int64) error {
	if b.commitErr != nil {
		return b.commitErr
	}

	b.next[partition] = next
	b.commits = append(b.commits, fmt.Sprintf("%d:%d", partition, next))

	return nil
}

func (b *fakeBackend) Publish(_ context.Context, message *sarama.ProducerMessage) error {
	if value, _ := message.Value.Encode(); b.failValue != "" && string(value) == b.failValue {
		return b.publishErr
	}

	b.published = append(b.published, message)

	return nil
}

func (b *fakeBackend) Close() error { return nil }

func TestList(t *testing.T) {
	ctx := context.Background()

	t.Run("lists the dead letters without moving the position", func(t *testing.T) {
		b := newBackend(2, 1)

		var entries []Entry
		count, err := List(ctx, b, "app_mail_dlq", 0, func(e Entry) { entries = append(entries, e) })
		if err != nil || count != 3 || len(entries) != 3 {
			t.Fatalf("count = %d, error = %v", count, err)
		}

		e := entries[1]
		want := Entry{
			Topic: "app_mail_dlq", Partition: 0, Offset: 1, Time: time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC),
			OriginalTopic: "app_mail", OriginalPartition: "1", OriginalOffset: "340", Group: "app_mail", Attempts: "3", Error: "smtp down",
			Key: []byte("key"), Value: []byte("body-0-1"),
		}
		if e.ID() != "app_mail_dlq[0]@1" || e.Topic != want.Topic || e.OriginalTopic != want.OriginalTopic || e.Error != want.Error ||
			e.Attempts != want.Attempts || e.Group != want.Group || e.OriginalPartition != "1" || e.OriginalOffset != "340" ||
			!e.Time.Equal(want.Time) || string(e.Key) != "key" || string(e.Value) != "body-0-1" {
			t.Errorf("entry = %+v, want %+v", e, want)
		}
		if len(b.commits) != 0 {
			t.Errorf("commits = %v, listing must not move the position", b.commits)
		}
	})

	t.Run("limit", func(t *testing.T) {
		b := newBackend(2, 2)

		var ids []string
		count, err := List(ctx, b, "app_mail_dlq", 3, func(e Entry) { ids = append(ids, e.ID()) })

		if err != nil || count != 3 || !slices.Equal(ids, []string{"app_mail_dlq[0]@0", "app_mail_dlq[0]@1", "app_mail_dlq[1]@0"}) {
			t.Errorf("count = %d, ids = %v, error = %v", count, ids, err)
		}
	})

	t.Run("starts from the saved position", func(t *testing.T) {
		b := newBackend(3)
		b.next[0] = 2

		count, _ := List(ctx, b, "app_mail_dlq", 0, func(Entry) {})
		if count != 1 {
			t.Errorf("count = %d, want the messages after the saved position", count)
		}
	})

	t.Run("failures", func(t *testing.T) {
		boom := errors.New("boom")

		b := newBackend(1)
		b.partitions = boom
		if _, err := List(ctx, b, "t", 0, func(Entry) {}); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}

		b = newBackend(1)
		b.position = boom
		if _, err := List(ctx, b, "t", 0, func(Entry) {}); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}
	})
}

func TestReplay(t *testing.T) {
	ctx := context.Background()

	t.Run("sends again to the original topic and saves the position", func(t *testing.T) {
		b := newBackend(2, 1)

		var replayed []string
		count, err := Replay(ctx, b, "app_mail_dlq", 0, func(e Entry) { replayed = append(replayed, e.ID()) })

		if err != nil || count != 3 || len(replayed) != 3 {
			t.Fatalf("count = %d, replayed = %v, error = %v", count, replayed, err)
		}
		if !slices.Equal(b.commits, []string{"0:1", "0:2", "1:1"}) {
			t.Errorf("commits = %v", b.commits)
		}

		m := b.published[0]
		value, _ := m.Value.Encode()
		key, _ := m.Key.Encode()
		if m.Topic != "app_mail" || string(value) != "body-0-0" || string(key) != "key" {
			t.Errorf("message = %+v", m)
		}
		if len(m.Headers) != 1 || string(m.Headers[0].Key) != HeaderReplayedFrom || string(m.Headers[0].Value) != "app_mail_dlq[0]@0" {
			t.Errorf("headers = %+v", m.Headers)
		}

		// Nothing is left to replay.
		if count, err := Replay(ctx, b, "app_mail_dlq", 0, func(Entry) {}); err != nil || count != 0 {
			t.Errorf("second replay: count = %d, error = %v", count, err)
		}
	})

	t.Run("limit", func(t *testing.T) {
		b := newBackend(3)

		count, err := Replay(ctx, b, "app_mail_dlq", 2, func(Entry) {})
		if err != nil || count != 2 || b.next[0] != 2 {
			t.Errorf("count = %d, position = %d, error = %v", count, b.next[0], err)
		}
	})

	t.Run("stops at the first failure and keeps it", func(t *testing.T) {
		boom := errors.New("boom")
		b := newBackend(3)
		b.failValue, b.publishErr = "body-0-1", boom

		count, err := Replay(ctx, b, "app_mail_dlq", 0, func(Entry) {})

		if !errors.Is(err, boom) || count != 1 {
			t.Errorf("count = %d, error = %v", count, err)
		}
		if b.next[0] != 1 {
			t.Errorf("position = %d, the failed message must stay to replay", b.next[0])
		}
	})

	t.Run("a message without original topic", func(t *testing.T) {
		b := newBackend(1)
		b.messages[0][0].Headers = nil

		if count, err := Replay(ctx, b, "app_mail_dlq", 0, func(Entry) {}); err == nil || count != 0 || len(b.published) != 0 {
			t.Errorf("count = %d, error = %v", count, err)
		}
	})

	t.Run("failure to save the position", func(t *testing.T) {
		boom := errors.New("boom")
		b := newBackend(1)
		b.commitErr = boom

		if _, err := Replay(ctx, b, "app_mail_dlq", 0, func(Entry) {}); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}
	})
}

func TestPurge(t *testing.T) {
	ctx := context.Background()

	t.Run("all", func(t *testing.T) {
		b := newBackend(2, 3)

		count, err := Purge(ctx, b, "app_mail_dlq", 0)

		if err != nil || count != 5 || !slices.Equal(b.commits, []string{"0:2", "1:3"}) || len(b.published) != 0 {
			t.Errorf("count = %d, commits = %v, error = %v", count, b.commits, err)
		}
	})

	t.Run("limit", func(t *testing.T) {
		b := newBackend(2, 3)

		count, err := Purge(ctx, b, "app_mail_dlq", 3)

		if err != nil || count != 3 || !slices.Equal(b.commits, []string{"0:2", "1:1"}) {
			t.Errorf("count = %d, commits = %v, error = %v", count, b.commits, err)
		}
	})

	t.Run("nothing to purge", func(t *testing.T) {
		b := newBackend(1)
		b.next[0] = 1

		if count, err := Purge(ctx, b, "app_mail_dlq", 0); err != nil || count != 0 || len(b.commits) != 0 {
			t.Errorf("count = %d, error = %v", count, err)
		}
	})

	t.Run("failures", func(t *testing.T) {
		boom := errors.New("boom")

		b := newBackend(1)
		b.partitions = boom
		if _, err := Purge(ctx, b, "t", 0); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}

		b = newBackend(1)
		b.position = boom
		if _, err := Purge(ctx, b, "t", 0); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}

		b = newBackend(1)
		b.commitErr = boom
		if _, err := Purge(ctx, b, "t", 0); !errors.Is(err, boom) {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("interrupted", func(t *testing.T) {
		canceled, cancel := context.WithCancel(ctx)
		cancel()

		if _, err := Purge(canceled, newBackend(1), "t", 0); !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v", err)
		}
	})
}
