package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/log"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// groupHandler handles the messages of the partitions of a topic given to a consumer of a group.
type groupHandler struct {
	topic  string
	group  string
	action func(value []byte) error

	// attempts is the number of times a message is handled, at least once, and delay the delay before the second one.
	attempts int
	delay    time.Duration

	// deadLetter receives a message which could not be handled.
	deadLetter func(message *sarama.ConsumerMessage, err error, attempts int)

	// stats counts the messages. It may be nil.
	stats *stats

	// dedup remembers the messages handled. It may be nil.
	dedup Deduplicator
}

// Setup is called when a consumer joins the group.
func (groupHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called when the consumer leaves the group or when the group is rebalanced.
func (groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim handles the messages of a partition one by one, and marks each as consumed once it is handled, or
// dead-lettered, so that a message is never lost if the process stops while it is handled.
func (h groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			// The body is not logged above the debug level: it holds secrets, such as the keys of the confirmation links.
			fields := logrus.Fields{
				"topic":     h.topic,
				"partition": message.Partition,
				"offset":    message.Offset,
				"size":      len(message.Value),
			}
			log.WithFields(fields).Info("kafka message received")
			log.WithFields(fields).WithField("message", string(message.Value)).Debug("kafka message body")
			h.stats.countReceived()

			if !h.handle(session.Context(), message) {
				return nil
			}
			session.MarkMessage(message, "")
			h.stats.setLag(message.Topic, message.Partition, claim.HighWaterMarkOffset()-message.Offset-1)

		case <-session.Context().Done():
			return nil
		}
	}
}

// handle handles the message, again if it fails, then dead-letters it if the last attempt fails or if the error is
// permanent. It reports false if ctx is done before the message is either handled or dead-lettered.
func (h groupHandler) handle(ctx context.Context, message *sarama.ConsumerMessage) bool {
	var err error

	id := messageID(message)
	if id != "" && h.dedup != nil && h.dedup.Handled(h.group, id) {
		h.stats.countDuplicate()
		log.WithFields(logrus.Fields{"topic": h.topic, "group": h.group, "message_id": id}).Warn("kafka message already handled, ignored")

		return true
	}

	attempt := 1
	for ; ; attempt++ {
		if err = h.try(message.Value); err == nil {
			h.stats.countHandled()
			if id != "" && h.dedup != nil {
				h.dedup.MarkHandled(h.group, id)
			}

			return true
		}

		fields := logrus.Fields{
			"topic":   h.topic,
			"group":   h.group,
			"attempt": attempt,
			"error":   err.Error(),
		}
		if IsPermanent(err) || attempt >= h.attempts {
			break
		}

		log.WithFields(fields).Warn("kafka message failed, trying again")
		h.stats.countRetried()
		if !sleep(ctx, h.backoff(attempt)) {
			return false
		}
	}

	if ctx.Err() != nil {
		return false
	}

	h.stats.countDeadLettered()
	h.deadLetter(message, err, attempt)

	return true
}

// try calls the action, and turns its panic into an error.
func (h groupHandler) try(value []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return h.action(value)
}

// backoff returns the delay to wait after the failure of an attempt: it doubles at each one.
func (h groupHandler) backoff(attempt int) time.Duration {
	delay := h.delay
	for range attempt - 1 {
		delay *= 2
		if delay >= maxRetryDelay {
			return maxRetryDelay
		}
	}

	return min(delay, maxRetryDelay)
}
