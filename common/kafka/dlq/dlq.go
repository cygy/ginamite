// Package dlq lists, replays and purges the dead letters: the messages the consumers of Kafka could not handle, see
// broker.MessageConsumer, which are kept in the topic of the dead letters of their topic, see kafka.DeadLetterTopic.
//
// The position of the tool in a topic is saved in Kafka, in a consumer group of its own: a message replayed or
// purged is never visited again, and List does not move it.
package dlq

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// Backend gives access to a topic of dead letters. Open returns the one of Kafka.
type Backend interface {
	// Partitions returns the partitions of the topic.
	Partitions(topic string) ([]int32, error)

	// Position returns the offset of the first message of the partition the tool did not visit yet, and the offset
	// after the last message.
	Position(topic string, partition int32) (next, end int64, err error)

	// Read calls fn with the messages of the partition whose offset is from up to to (excluded), in order. It stops
	// at the first error of fn.
	Read(ctx context.Context, topic string, partition int32, from, to int64, fn func(*sarama.ConsumerMessage) error) error

	// Commit saves that the messages of the partition before the offset are visited.
	Commit(topic string, partition int32, next int64) error

	// Publish sends a message, and waits for its acknowledgement.
	Publish(ctx context.Context, message *sarama.ProducerMessage) error

	// Close releases the connections.
	Close() error
}

// Names of the headers of the dead letters, set by the consumers, and of the messages replayed.
const (
	headerError     = "error"
	headerAttempts  = "attempts"
	headerGroup     = "group"
	headerTopic     = "topic"
	headerPartition = "partition"
	headerOffset    = "offset"

	// HeaderReplayedFrom is the header of a replayed message, which gives where the dead letter was.
	HeaderReplayedFrom = "replayed-from"
)

// Entry is a dead letter.
type Entry struct {
	Topic     string // The topic of the dead letters.
	Partition int32
	Offset    int64
	Time      time.Time

	OriginalTopic     string // The topic the message was consumed from, with the prefix.
	OriginalPartition string
	OriginalOffset    string
	Group             string
	Attempts          string
	Error             string

	Key, Value []byte
}

// ID returns where the dead letter is: "topic[partition]@offset".
func (e Entry) ID() string {
	return fmt.Sprintf("%s[%d]@%d", e.Topic, e.Partition, e.Offset)
}

func newEntry(message *sarama.ConsumerMessage) Entry {
	e := Entry{
		Topic:     message.Topic,
		Partition: message.Partition,
		Offset:    message.Offset,
		Time:      message.Timestamp,
		Key:       message.Key,
		Value:     message.Value,
	}

	for _, header := range message.Headers {
		value := string(header.Value)

		switch string(header.Key) {
		case headerError:
			e.Error = value
		case headerAttempts:
			e.Attempts = value
		case headerGroup:
			e.Group = value
		case headerTopic:
			e.OriginalTopic = value
		case headerPartition:
			e.OriginalPartition = value
		case headerOffset:
			e.OriginalOffset = value
		}
	}

	return e
}

// List calls visit with the dead letters not replayed nor purged yet, partition by partition, up to limit of them, or
// all of them if limit is 0. It returns their count.
func List(ctx context.Context, b Backend, topic string, limit int, visit func(Entry)) (int, error) {
	return walk(ctx, b, topic, limit, func(e Entry) error {
		visit(e)

		return nil
	}, false)
}

// Replay sends again, to the topic they were consumed from, the dead letters not replayed nor purged yet, partition by
// partition, up to limit of them, or all of them if limit is 0, and calls replayed with each one. A dead letter is
// marked as visited once it is acknowledged. It stops at the first failure, which it returns, with the count of the
// dead letters replayed: they can be replayed later.
func Replay(ctx context.Context, b Backend, topic string, limit int, replayed func(Entry)) (int, error) {
	return walk(ctx, b, topic, limit, func(e Entry) error {
		if e.OriginalTopic == "" {
			return fmt.Errorf("%s has no original topic", e.ID())
		}

		message := &sarama.ProducerMessage{
			Topic:   e.OriginalTopic,
			Key:     sarama.ByteEncoder(e.Key),
			Value:   sarama.ByteEncoder(e.Value),
			Headers: []sarama.RecordHeader{{Key: []byte(HeaderReplayedFrom), Value: []byte(e.ID())}},
		}
		if err := b.Publish(ctx, message); err != nil {
			return fmt.Errorf("unable to send %s again: %w", e.ID(), err)
		}

		replayed(e)

		return nil
	}, true)
}

// Purge marks as visited, without sending them again, the dead letters not replayed nor purged yet, partition by
// partition, up to limit of them, or all of them if limit is 0. It returns their count.
func Purge(ctx context.Context, b Backend, topic string, limit int) (int, error) {
	partitions, err := b.Partitions(topic)
	if err != nil {
		return 0, fmt.Errorf("unable to get the partitions of %s: %w", topic, err)
	}

	purged := 0
	for _, partition := range partitions {
		if err := ctx.Err(); err != nil {
			return purged, err
		}

		next, end, err := b.Position(topic, partition)
		if err != nil {
			return purged, err
		}

		to := end
		if limit > 0 {
			to = min(end, next+int64(limit-purged))
		}
		if to <= next {
			continue
		}

		if err := b.Commit(topic, partition, to); err != nil {
			return purged, fmt.Errorf("unable to save the position in %s[%d]: %w", topic, partition, err)
		}
		purged += int(to - next)

		if limit > 0 && purged >= limit {
			break
		}
	}

	return purged, nil
}

// walk calls visit with the dead letters not visited yet, up to limit of them, and saves the position after each
// one if commit is true. It stops at the first error of visit.
func walk(ctx context.Context, b Backend, topic string, limit int, visit func(Entry) error, commit bool) (int, error) {
	partitions, err := b.Partitions(topic)
	if err != nil {
		return 0, fmt.Errorf("unable to get the partitions of %s: %w", topic, err)
	}

	count := 0
	for _, partition := range partitions {
		next, end, err := b.Position(topic, partition)
		if err != nil {
			return count, err
		}
		if limit > 0 {
			end = min(end, next+int64(limit-count))
		}
		if end <= next {
			continue
		}

		err = b.Read(ctx, topic, partition, next, end, func(message *sarama.ConsumerMessage) error {
			if err := visit(newEntry(message)); err != nil {
				return err
			}
			count++

			if !commit {
				return nil
			}

			if err := b.Commit(topic, partition, message.Offset+1); err != nil {
				return fmt.Errorf("unable to save the position in %s[%d]: %w", topic, partition, err)
			}

			return nil
		})
		if err != nil {
			return count, err
		}

		if limit > 0 && count >= limit {
			break
		}
	}

	return count, nil
}
