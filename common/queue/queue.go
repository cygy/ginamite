package queue

import (
	"github.com/cygy/ginamite/common/log"
)

// Producer sends a message to a topic. The Kafka client of the package kafka is a Producer.
type Producer interface {
	SendMessageToTopic(message, topic string)
}

// Queue sends the tasks, the cache messages, the mails and the notifications to the other servers, through a Producer.
type Queue struct {
	producer Producer
}

// New returns a Queue sending its messages with the producer.
func New(producer Producer) *Queue {
	return &Queue{producer: producer}
}

// send sends the message to the topic. It logs an error and does nothing if the queue has no producer.
func (q *Queue) send(message, topic string) {
	if q == nil || q.producer == nil {
		log.WithField("topic", topic).Error("unable to send the message: the queue has no producer")
		return
	}

	q.producer.SendMessageToTopic(message, topic)
}
