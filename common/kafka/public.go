package kafka

import "github.com/cygy/ginamite/common/log"

// Initialize : initializes the kafka producer/consumer.
func Initialize(host string, port int, prefix string, hasProducer bool, consumers []MessageConsumer) {
	Main = NewClient()
	Main.Initialize(host, port, prefix, hasProducer, consumers)
}

// SendMessageToTopic : sends a message to a topic.
func SendMessageToTopic(message, topic string) {
	Main.SendMessageToTopic(message, topic)
}

// Recover : recovers from a panic, use this function to the consumers.
func Recover() {
	if r := recover(); r != nil {
		log.WithField("panic", r).Warning("recover from panic")
	}
}
