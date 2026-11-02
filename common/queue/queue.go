package queue

import (
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/log"
)

// Keyed is implemented by the payloads which have a key: the messages of the same key are consumed in the order they
// are sent, such as the messages of a same user.
type Keyed interface {
	PartitionKey() string
}

// Queue sends the tasks, the cache messages, the mails and the notifications to the other servers, through a broker.Producer.
type Queue struct {
	producer broker.Producer
}

// New returns a Queue sending its messages with the producer.
func New(producer broker.Producer) *Queue {
	return &Queue{producer: producer}
}

// send sends the message to the topic, with the key if it is not empty and the producer is a KeyedProducer. It logs an
// error and does nothing if the queue has no producer.
func (q *Queue) send(key, message, topic string) {
	if q == nil || q.producer == nil {
		log.Error("unable to send the message: the queue has no producer", "topic", topic)
		return
	}

	if keyed, ok := q.producer.(broker.KeyedProducer); ok && key != "" {
		keyed.SendKeyedMessageToTopic(key, message, topic)
		return
	}

	q.producer.SendMessageToTopic(message, topic)
}
