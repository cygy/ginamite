package kafka

import "github.com/cygy/ginamite/common/log"

// Main is the main Kafka client of the application, set by Initialize.
var Main *Client

// Initialize creates the main client Main and initializes it.
// See Client.Initialize.
func Initialize(host string, port int, prefix string, hasProducer bool, consumers []MessageConsumer) {
	Main = NewClient()
	Main.Initialize(host, port, prefix, hasProducer, consumers)
}

// SendMessageToTopic sends a message to the topic with the main client.
// See Client.SendMessageToTopic.
func SendMessageToTopic(message, topic string) {
	Main.SendMessageToTopic(message, topic)
}

// Recover recovers from a panic and logs it as a warning.
// Defer it in the message consumers so that a panic does not stop them.
func Recover() {
	if r := recover(); r != nil {
		log.WithField("panic", r).Warning("recover from panic")
	}
}
