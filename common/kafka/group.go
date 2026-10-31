package kafka

import (
	"github.com/cygy/ginamite/common/log"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// groupHandler handles the messages of the partitions of a topic given to a consumer of a group.
type groupHandler struct {
	topic  string
	action func(value []byte)
}

// Setup is called when a consumer joins the group.
func (groupHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called when the consumer leaves the group or when the group is rebalanced.
func (groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim handles the messages of a partition one by one, and marks each as consumed once it is handled, so that
// a message is never lost if the process stops while it is handled.
func (h groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			log.WithFields(logrus.Fields{
				"topic":   h.topic,
				"message": string(message.Value),
			}).Info("kafka message received")

			h.action(message.Value)
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
