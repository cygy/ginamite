package worker

import (
	"context"

	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/mail"
	"github.com/cygy/ginamite/worker/notification"
	"github.com/cygy/ginamite/worker/tasks/async/location"
	"github.com/cygy/ginamite/worker/tasks/async/user"

	"github.com/sirupsen/logrus"
)

// CreateKafkaConsumers returns the consumers of the topics of the tasks, of the group notifications,
// of the user notifications and of the mails. A panic of a consumer is recovered.
func (s *Server) CreateKafkaConsumers() []kafka.MessageConsumer {
	return []kafka.MessageConsumer{
		{Topic: queue.TopicTasks, Action: s.handleTask},
		{Topic: queue.TopicGroupNotification, Action: s.handleGroupNotification},
		{Topic: queue.TopicUserNotification, Action: s.handleUserNotification},
		{Topic: queue.TopicMail, Action: s.handleMail},
	}
}

// handleTask handles a message of the topic of the tasks.
func (s *Server) handleTask(value []byte) {
	defer kafka.Recover()

	msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicTasks)
	if !ok {
		return
	}

	switch msgType {
	case queue.MessageTaskIPLocation:
		location.GetIPInfo(payload, s.Functions.SaveIPAddressDetails)
	case queue.MessageTaskDeleteUser:
		user.DeleteByID(payload, s.Functions.DeleteUserByID)
	case queue.MessageTaskDisableUser:
		user.DisableByID(payload, s.Functions.DisableUserByID)
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
}

// handleGroupNotification handles a message of the topic of the group notifications.
func (s *Server) handleGroupNotification(value []byte) {
	defer kafka.Recover()

	msgType, payloadWithAbility, ok := queue.ParseMessageAndPayload(value, queue.TopicGroupNotification)
	if !ok {
		return
	}

	ability, payload, ok := queue.ParseGroupAbilityAndPayload(payloadWithAbility, queue.TopicGroupNotification)
	if !ok {
		return
	}

	if s.Functions.HandleNotificationForUsersWithAbility == nil || !s.Functions.HandleNotificationForUsersWithAbility(context.Background(), ability, msgType, payload) {
		warnUnknownMessage(queue.TopicGroupNotification, msgType)
	}
}

// handleUserNotification handles a message of the topic of the user notifications.
func (s *Server) handleUserNotification(value []byte) {
	defer kafka.Recover()

	msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicUserNotification)
	if !ok {
		return
	}

	switch msgType {
	case notifications.TypeNewLogin:
		notification.NewLogin(context.Background(), payload, s.Functions.GetNotificationTargetsByUserAndType)
	default:
		if s.Functions.HandleUserNotification == nil || !s.Functions.HandleUserNotification(context.Background(), msgType, payload, s.Functions.GetNotificationTargetsByUserAndType) {
			warnUnknownMessage(queue.TopicUserNotification, msgType)
		}
	}
}

// handleMail handles a message of the topic of the mails.
func (s *Server) handleMail(value []byte) {
	defer kafka.Recover()

	msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicMail)
	if !ok {
		return
	}

	if mail.Handle(context.Background(), msgType, payload, s.Functions.IsEmailAddressValid) {
		return
	}

	if s.Functions.HandleMail == nil || !s.Functions.HandleMail(context.Background(), msgType, payload, s.Functions.IsEmailAddressValid) {
		warnUnknownMessage(queue.TopicMail, msgType)
	}
}

// warnUnknownMessage logs a warning about a message no function handled.
func warnUnknownMessage(topic, msgType string) {
	log.WithFields(logrus.Fields{
		"topic": topic,
		"type":  msgType,
	}).Warn("unknown kafka message")
}
