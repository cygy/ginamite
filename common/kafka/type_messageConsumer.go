package kafka

// MessageConsumer : a message consumer consumes messages from a queue 'topic' and executes 'action'
type MessageConsumer struct {
	Topic  string
	Action func(value []byte)
}
