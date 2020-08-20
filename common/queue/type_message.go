package queue

import (
	"encoding/json"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// Message : struct Message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Decode : decodes a message received.
func (m *Message) Decode(bytes []byte) bool {
	err := json.Unmarshal(bytes, m)

	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("unable to decode kafka message")
	}

	return err == nil
}
