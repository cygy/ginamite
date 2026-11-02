package worker

import (
	"context"
	"errors"
	"slices"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/internal/async/user"
	"github.com/cygy/ginamite/worker/internal/notification"
	"github.com/cygy/ginamite/worker/tasks/async/location"
)

// errUndecodable is the error of a message which can not be decoded: it is dead-lettered without being tried again.
var errUndecodable = broker.Permanent(errors.New("undecodable message"))

// CreateConsumers returns the consumers of the topics of the tasks (custom, location, user and registration), of the group notifications,
// of the user notifications and of the mails. A message which can not be decoded is dead-lettered, and so is a message
// whose handling failed, after some attempts. A panic is handled like an error.
//
// A worker consumes all the topics, unless the configuration lists the ones it consumes, with their consumer group:
// the workers of the same group share the messages of the topic, so several workers can handle the same topic, and
// each handle different ones.
func (s *Server) CreateConsumers() []broker.MessageConsumer {
	all := []broker.MessageConsumer{
		{Topic: queue.TopicTasks, Action: s.handleTask(queue.TopicTasks)},
		{Topic: queue.TopicTaskLocation, Action: s.handleTask(queue.TopicTaskLocation)},
		{Topic: queue.TopicTaskUser, Action: s.handleTask(queue.TopicTaskUser)},
		{Topic: queue.TopicTaskRegistration, Action: s.handleTask(queue.TopicTaskRegistration)},
		{Topic: queue.TopicGroupNotification, Action: s.handleGroupNotification},
		{Topic: queue.TopicUserNotification, Action: s.handleUserNotification},
		{Topic: queue.TopicMail, Action: s.handleMail},
	}
	if s.Config == nil || len(s.Config.WorkerConsumers) == 0 {
		return all
	}

	consumers := make([]broker.MessageConsumer, 0, len(s.Config.WorkerConsumers))
	for _, wanted := range s.Config.WorkerConsumers {
		i := slices.IndexFunc(all, func(c broker.MessageConsumer) bool { return c.Topic == wanted.Topic })
		if i < 0 {
			log.Error("unknown kafka topic of the worker", "topic", wanted.Topic)
			continue
		}

		consumer := all[i]
		consumer.Group = wanted.Group
		consumers = append(consumers, consumer)
	}

	return consumers
}

// handleTask returns the handler of the messages of a topic of the tasks. All the topics handle the built-in tasks,
// so the ones sent to the topic "task" before they had their own topic are still handled, then the custom tasks.
func (s *Server) handleTask(topic string) func(value []byte) error {
	return func(value []byte) error {
		return s.handleTaskMessage(topic, value)
	}
}

// handleTaskMessage handles a message of the topic of tasks.
func (s *Server) handleTaskMessage(topic string, value []byte) error {
	msgType, payload, ok := queue.ParseMessageAndPayload(value, topic)
	if !ok {
		return errUndecodable
	}

	switch msgType {
	case queue.MessageTaskIPLocation:
		return location.GetIPInfo(payload, s.Config.IPStack, s.Storage.saveIPAddressDetailsFunc())
	case queue.MessageTaskDeleteUser:
		return user.DeleteByID(payload, s.Queue, s.Storage.deleteUserFunc())
	case queue.MessageTaskDisableUser:
		return user.DisableByID(payload, s.Queue, s.Storage.disableUserFunc())
	case queue.MessageTaskEnableUser:
		return user.EnableByID(payload, s.Storage.enableUserFunc())
	case queue.MessageTaskUpdateUserSocialNetworks:
		return user.UpdateSocialNetworksByID(payload, s.Storage.updateSocialNetworksFunc())
	case queue.MessageTaskRegistrationDone:
		return user.RegistrationDone(payload, s.Hooks.RegistrationDone)
	case queue.MessageTaskRegistrationValidated:
		return user.RegistrationValidated(payload, s.Hooks.RegistrationValidated)
	}

	if s.Hooks.HandleTask != nil {
		handled, err := s.Hooks.HandleTask(context.Background(), msgType, payload)
		if handled {
			return err
		}
	}
	warnUnknownMessage(topic, msgType)

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

	if s.Hooks.HandleNotificationForUsersWithAbility != nil {
		handled, err := s.Hooks.HandleNotificationForUsersWithAbility(context.Background(), ability, msgType, payload)
		if handled {
			return err
		}
	}
	warnUnknownMessage(queue.TopicGroupNotification, msgType)

	return nil
}

// handleUserNotification handles a message of the topic of the user notifications.
func (s *Server) handleUserNotification(value []byte) error {
	msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicUserNotification)
	if !ok {
		return errUndecodable
	}

	if msgType == notifications.TypeNewLogin {
		return notification.NewLogin(context.Background(), s.Queue, payload, s.Storage.targetsFunc())
	}

	if s.Hooks.HandleUserNotification != nil {
		handled, err := s.Hooks.HandleUserNotification(context.Background(), msgType, payload, s.Storage.targetsFunc())
		if handled {
			return err
		}
	}
	warnUnknownMessage(queue.TopicUserNotification, msgType)

	return nil
}

// handleMail handles a message of the topic of the mails.
func (s *Server) handleMail(value []byte) error {
	msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicMail)
	if !ok {
		return errUndecodable
	}

	if handled, err := s.mailer.Handle(context.Background(), msgType, payload, s.Storage.emailCheckFunc()); handled {
		return err
	}

	if s.Hooks.HandleMail != nil {
		handled, err := s.Hooks.HandleMail(context.Background(), msgType, payload, s.Storage.emailCheckFunc())
		if handled {
			return err
		}
	}
	warnUnknownMessage(queue.TopicMail, msgType)

	return nil
}

// warnUnknownMessage logs a warning about a message no function handled.
func warnUnknownMessage(topic, msgType string) {
	log.Warn("unknown kafka message",
		"topic", topic,
		"type", msgType,
	)
}
