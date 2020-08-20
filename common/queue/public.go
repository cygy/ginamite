package queue

import (
	"encoding/json"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// CreateTask : sends a new task to the topic 'task'.
func CreateTask(taskName string, payload interface{}) {
	sendMessage(taskName, payload, TopicTasks)
}

// ParseMessageAndPayload : parses the content of a kafka message.
func ParseMessageAndPayload(value []byte, topic string) (string, []byte, bool) {
	msg := &Message{}
	if ok := msg.Decode(value); !ok {
		log.WithFields(logrus.Fields{
			"topic": topic,
		}).Error("unable to decode the message")
		return "", nil, false
	}

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		log.WithFields(logrus.Fields{
			"topic": topic,
			"type":  msg.Type,
		}).Error("unable to parse the payload from json")
		return "", nil, false
	}

	return msg.Type, payload, true
}

// ParseGroupAbilityAndPayload : parses the content of a kafka message.
func ParseGroupAbilityAndPayload(value []byte, topic string) (string, []byte, bool) {
	msg := &GroupNotification{}
	if err := json.Unmarshal(value, msg); err != nil {
		log.WithFields(logrus.Fields{
			"topic": topic,
			"error": err.Error(),
		}).Error("unable to decode kafka message")
		return "", nil, false
	}

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		log.WithFields(logrus.Fields{
			"topic": topic,
			"type":  msg.Ability,
			"error": err.Error(),
		}).Error("unable to parse the payload from json")
		return "", nil, false
	}

	return msg.Ability, payload, true
}

// UnmarshalPayload : unmarshals a payload of a kafka message.
func UnmarshalPayload(data []byte, v interface{}) error {
	err := json.Unmarshal(data, &v)
	if err != nil {
		log.Error("unable to parse the payload from json")
	}

	return err
}
