package queue

import (
	"encoding/json"

	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

func sendMessage(messageType string, payload interface{}, topic string) {
	message := Message{
		Type:    messageType,
		Payload: payload,
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		log.WithFields(logrus.Fields{
			"type":    messageType,
			"payload": payload,
		}).Error("unable to marshal JSON from message")
		return
	}

	kafka.SendMessageToTopic(string(bytes), topic)
}
