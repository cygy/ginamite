package queue

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// Message is a message exchanged through Kafka: a type and its payload.
type Message struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// Decode decodes the JSON content of a received message into m.
// It logs an error and returns false if the content is invalid.
func (m *Message) Decode(bytes []byte) bool {
	err := json.Unmarshal(bytes, m)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("unable to decode kafka message")
	}

	return err == nil
}

// ParseMessageAndPayload decodes a message received from the topic and returns its
// type and its payload as JSON, ready to be unmarshaled by UnmarshalPayload.
// It logs an error and returns false if the message can not be decoded.
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

// ParseGroupAbilityAndPayload decodes a group notification received from the topic
// and returns its ability and its payload as JSON.
// It logs an error and returns false if the message can not be decoded.
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

// UnmarshalPayload unmarshals the JSON payload of a message into v, which must be a pointer.
// It logs an error if the payload is invalid.
func UnmarshalPayload(data []byte, v any) error {
	err := json.Unmarshal(data, &v)
	if err != nil {
		log.Error("unable to parse the payload from json")
	}

	return err
}

// sendMessage sends a message of the given type and payload to the topic.
func (q *Queue) sendMessage(messageType string, payload any, topic string) {
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

	q.send(partitionKey(payload), string(bytes), topic)
}

// partitionKeyFields are the JSON names of the fields identifying the entity a payload is about, by priority.
var partitionKeyFields = []string{"user_id", "email"}

// partitionKey returns the key of the message of the payload, so the messages about the same entity are consumed in the
// order they are sent. It is the PartitionKey of a Keyed payload, else the value of the first field of a struct
// payload named user_id or email in its JSON tag. It is empty if the payload has no key.
func partitionKey(payload any) string {
	if keyed, ok := payload.(Keyed); ok {
		return keyed.PartitionKey()
	}

	value := reflect.ValueOf(payload)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return ""
	}

	for _, name := range partitionKeyFields {
		for i := range value.NumField() {
			tag, _, _ := strings.Cut(value.Type().Field(i).Tag.Get("json"), ",")
			if field := value.Field(i); tag == name && field.Kind() == reflect.String && field.String() != "" {
				return field.String()
			}
		}
	}

	return ""
}
