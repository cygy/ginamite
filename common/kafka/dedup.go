package kafka

import "github.com/IBM/sarama"

// HeaderMessageID is the header holding the identifier of a message, set by SendKeyedMessageToTopic.
const HeaderMessageID = "message-id"

// messageID returns the identifier of the message, or "" if it has none.
func messageID(message *sarama.ConsumerMessage) string {
	for _, header := range message.Headers {
		if header != nil && string(header.Key) == HeaderMessageID {
			return string(header.Value)
		}
	}

	return ""
}
