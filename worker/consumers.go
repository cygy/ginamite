package worker

import (
	"context"
	"errors"
	"slices"

	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/notification"
	"github.com/cygy/ginamite/worker/tasks/async/location"
	"github.com/cygy/ginamite/worker/tasks/async/user"

	"github.com/sirupsen/logrus"
)

// errUndecodable is the error of a message which can not be decoded: it is dead-lettered without being tried again.
var errUndecodable = kafka.Permanent(errors.New("undecodable message"))

// CreateKafkaConsumers returns the consumers of the topics of the tasks, of the group notifications,
// of the user notifications and of the mails. A message which can not be decoded, or whose email can not be sent, is
// dead-lettered, after some attempts for the second case. A panic is handled like an error.
//
// A worker consumes all the topics, unless the configuration lists the ones it consumes, with their consumer group:
// the workers of the same group share the messages of the topic, so several workers can handle the same topic, and
// each handle different ones.
func (s *Server) CreateKafkaConsumers() []kafka.MessageConsumer {
	all := []kafka.MessageConsumer{
		{Topic: queue.TopicTasks, Action: s.handleTask},
		{Topic: queue.TopicGroupNotification, Action: s.handleGroupNotification},
		{Topic: queue.TopicUserNotification, Action: s.handleUserNotification},
		{Topic: queue.TopicMail, Action: s.handleMail},
	}
	if s.Config == nil || len(s.Config.WorkerConsumers) == 0 {
		return all
	}

	consumers := make([]kafka.MessageConsumer, 0, len(s.Config.WorkerConsumers))
	for _, wanted := range s.Config.WorkerConsumers {
		i := slices.IndexFunc(all, func(c kafka.MessageConsumer) bool { return c.Topic == wanted.Topic })
		if i < 0 {
			log.WithField("topic", wanted.Topic).Error("unknown kafka topic of the worker")
			continue
		}

		consumer := all[i]
		consumer.Group = wanted.Group
		consumers = append(consumers, consumer)
	}

	return consumers
}

// handleTask handles a message of the topic of the tasks.
func (s *Server) handleTask(value []byte) error {
	msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicTasks)
	if !ok {
		return errUndecodable
	}

	switch msgType {
	case queue.MessageTaskIPLocation:
		location.GetIPInfo(payload, s.Config.IPStack, s.Functions.SaveIPAddressDetails)
	case queue.MessageTaskDeleteUser:
		user.DeleteByID(payload, s.Queue, s.Functions.DeleteUserByID)
	case queue.MessageTaskDisableUser:
		user.DisableByID(payload, s.Queue, s.Functions.DisableUserByID)
	case queue.MessageTaskEnableUser:
		user.EnableByID(payload, s.Functions.EnableUserByID)
	case queue.MessageTaskUpdateUserSocialNetworks:
		user.UpdateSocialNetworksByID(payload, s.Functions.UpdateUserSocialNetworksByID)
	case queue.MessageTaskRegistrationDone:
		user.RegistrationDone(payload, s.Functions.RegistrationDone)
	case queue.MessageTaskRegistrationValidated:
		user.RegistrationValidated(payload, s.Functions.RegistrationValidated)
	default:
		if s.Functions.HandleTask == nil || !s.Functions.HandleTask(context.Background(), msgType, payload) {
			warnUnknownMessage(queue.TopicTasks, msgType)
		}
	}

	return nil
}

// handleGroupNotification handles a message of the topic of the group notifications.
func (s *Server) handleGroupNotification(value []byte) error {
	msgType, payloadWithAbility, ok := queue.ParseMessageAndPayload(value, queue.TopicGroupNotification)
	if !ok {
		return errUndecodable
	}

	ability, payload, ok := queue.ParseGroupAbilityAndPayload(payloadWithAbility, queue.TopicGroupNotification)
	if !ok {
		return errUndecodable
	}

	if s.Functions.HandleNotificationForUsersWithAbility == nil || !s.Functions.HandleNotificationForUsersWithAbility(context.Background(), ability, msgType, payload) {
		warnUnknownMessage(queue.TopicGroupNotification, msgType)
	}

	return nil
}

// handleUserNotification handles a message of the topic of the user notifications.
func (s *Server) handleUserNotification(value []byte) error {
	msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicUserNotification)
	if !ok {
		return errUndecodable
	}

	switch msgType {
	case notifications.TypeNewLogin:
		notification.NewLogin(context.Background(), s.Queue, payload, s.Functions.GetNotificationTargetsByUserAndType)
	default:
		if s.Functions.HandleUserNotification == nil || !s.Functions.HandleUserNotification(context.Background(), msgType, payload, s.Functions.GetNotificationTargetsByUserAndType) {
			warnUnknownMessage(queue.TopicUserNotification, msgType)
		}
	}

	return nil
}

// handleMail handles a message of the topic of the mails.
func (s *Server) handleMail(value []byte) error {
	msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicMail)
	if !ok {
		return errUndecodable
	}

	if handled, err := s.mailer.Handle(context.Background(), msgType, payload, s.Functions.IsEmailAddressValid); handled {
		return err
	}

	if s.Functions.HandleMail == nil || !s.Functions.HandleMail(context.Background(), msgType, payload, s.Functions.IsEmailAddressValid) {
		warnUnknownMessage(queue.TopicMail, msgType)
	}

	return nil
}

// warnUnknownMessage logs a warning about a message no function handled.
func warnUnknownMessage(topic, msgType string) {
	log.WithFields(logrus.Fields{
		"topic": topic,
		"type":  msgType,
	}).Warn("unknown kafka message")
}
