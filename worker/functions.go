package worker

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/template/text"
	"github.com/cygy/ginamite/common/weburl"
	"github.com/cygy/ginamite/worker/mail"
	"github.com/cygy/ginamite/worker/notification"
	"github.com/cygy/ginamite/worker/tasks/async/location"
	"github.com/cygy/ginamite/worker/tasks/async/user"
	"github.com/cygy/ginamite/worker/tasks/recurring"
	"github.com/cygy/ginamite/worker/tasks/recurring/account"
	"github.com/cygy/ginamite/worker/tasks/recurring/authtokens"
	"github.com/cygy/ginamite/worker/tasks/recurring/sitemap"

	"github.com/sirupsen/logrus"
)

// CreateKafkaConsumers return an array of kafka consumers.
func (s *Server) CreateKafkaConsumers() []kafka.MessageConsumer {
	consumers := []kafka.MessageConsumer{}

	// Handle the tasks.
	consumers = append(consumers, kafka.MessageConsumer{
		Topic: queue.TopicTasks,
		Action: func(value []byte) {
			defer kafka.Recover()

			msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicTasks)
			if !ok {
				return
			}

			switch msgType {
			case queue.MessageTaskIPLocation:
				location.GetIPInfo(payload, s.Functions.SaveIPAddressDetails)
				break
			case queue.MessageTaskDeleteUser:
				user.DeleteByID(payload, s.Functions.DeleteUserByID)
				break
			case queue.MessageTaskDisableUser:
				user.DisableByID(payload, s.Functions.DisableUserByID)
				break
			case queue.MessageTaskEnableUser:
				user.EnableByID(payload, s.Functions.EnableUserByID)
				break
			case queue.MessageTaskUpdateUserSocialNetworks:
				user.UpdateSocialNetworksByID(payload, s.Functions.UpdateUserSocialNetworksByID)
				break
			case queue.MessageTaskRegistrationDone:
				user.RegistrationDone(payload, s.Functions.RegistrationDone)
				break
			case queue.MessageTaskRegistrationValidated:
				user.RegistrationDone(payload, s.Functions.RegistrationValidated)
				break
			default:
				if s.Functions.HandleTask == nil || !s.Functions.HandleTask(msgType, payload) {
					log.WithFields(logrus.Fields{
						"topic": queue.TopicTasks,
						"type":  msgType,
					}).Warn("unknown kafka message")
				}
			}
		}})

	// Handle the group notifications to send.
	consumers = append(consumers, kafka.MessageConsumer{
		Topic: queue.TopicGroupNotification,
		Action: func(value []byte) {
			defer kafka.Recover()

			msgType, payloadWithAbility, ok := queue.ParseMessageAndPayload(value, queue.TopicGroupNotification)
			if !ok {
				return
			}

			ability, payload, ok := queue.ParseGroupAbilityAndPayload(payloadWithAbility, queue.TopicGroupNotification)
			if !ok {
				return
			}

			if s.Functions.HandleNotificationForUsersWithAbility == nil || !s.Functions.HandleNotificationForUsersWithAbility(ability, msgType, payload) {
				log.WithFields(logrus.Fields{
					"topic": queue.TopicGroupNotification,
					"type":  msgType,
				}).Warn("unknown kafka message")
			}
		}})

	// Handle the notifications to send.
	consumers = append(consumers, kafka.MessageConsumer{
		Topic: queue.TopicUserNotification,
		Action: func(value []byte) {
			defer kafka.Recover()

			msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicUserNotification)
			if !ok {
				return
			}

			switch msgType {
			case notifications.TypeNewLogin:
				notification.NewLogin(payload, s.Functions.GetNotificationTargetsByUserAndType)
				break
			default:
				if s.Functions.HandleUserNotification == nil || !s.Functions.HandleUserNotification(msgType, payload, s.Functions.GetNotificationTargetsByUserAndType) {
					log.WithFields(logrus.Fields{
						"topic": queue.TopicUserNotification,
						"type":  msgType,
					}).Warn("unknown kafka message")
				}
			}
		}})

	// Handle the emails to send.
	consumers = append(consumers, kafka.MessageConsumer{
		Topic: queue.TopicMail,
		Action: func(value []byte) {
			defer kafka.Recover()

			isEmailAddressValid := s.Functions.IsEmailAddressValid

			msgType, payload, ok := queue.ParseMessageAndPayload(value, queue.TopicMail)
			if !ok {
				return
			}

			switch msgType {
			case queue.MessageMailRegistrationConfirmation:
				mail.RegistrationConfirmation(payload)
				break
			case queue.MessageMailRegistrationConfirmed:
				mail.RegistrationConfirmed(payload)
				break
			case queue.MessageMailRegistrationCancelled:
				mail.RegistrationCancelled(payload)
				break
			case queue.MessageMailRegistrationWelcome:
				mail.RegistrationWelcome(payload)
				break
			case queue.MessageMailForgotPasswordConfirmation:
				mail.ForgotPasswordConfirmation(payload, isEmailAddressValid)
				break
			case queue.MessageMailForgotPasswordConfirmed:
				mail.ForgotPasswordConfirmed(payload, isEmailAddressValid)
				break
			case queue.MessageMailForgotPasswordCancelled:
				mail.ForgotPasswordCancelled(payload, isEmailAddressValid)
				break
			case queue.MessageMailUpdateEmailAddressConfirmation:
				mail.UpdateEmailAddressConfirmation(payload, isEmailAddressValid)
				break
			case queue.MessageMailUpdateEmailAddressConfirmed:
				mail.UpdateEmailAddressConfirmed(payload, isEmailAddressValid)
				break
			case queue.MessageMailUpdateEmailAddressCancelled:
				mail.UpdateEmailAddressCancelled(payload, isEmailAddressValid)
				break
			case queue.MessageMailUpdatePasswordConfirmation:
				mail.UpdatePasswordConfirmation(payload, isEmailAddressValid)
				break
			case queue.MessageMailUpdatePasswordConfirmed:
				mail.UpdatePasswordConfirmed(payload, isEmailAddressValid)
				break
			case queue.MessageMailUpdatePasswordCancelled:
				mail.UpdatePasswordCancelled(payload, isEmailAddressValid)
				break
			case queue.MessageMailValidateNewEmailAddress:
				mail.ValidateNewEmailAddress(payload, isEmailAddressValid)
				break
			case queue.MessageMailDeleteAccountConfirmation:
				mail.DeleteAccountConfirmation(payload, isEmailAddressValid)
				break
			case queue.MessageMailDeleteAccountConfirmed:
				mail.DeleteAccountConfirmed(payload, isEmailAddressValid)
				break
			case queue.MessageMailDeleteAccountCancelled:
				mail.DeleteAccountCancelled(payload, isEmailAddressValid)
				break
			case queue.MessageMailDisableAccountConfirmation:
				mail.DisableAccountConfirmation(payload, isEmailAddressValid)
				break
			case queue.MessageMailDisableAccountConfirmed:
				mail.DisableAccountConfirmed(payload, isEmailAddressValid)
				break
			case queue.MessageMailDisableAccountCancelled:
				mail.DisableAccountCancelled(payload, isEmailAddressValid)
				break
			case queue.MessageMailEnableAccountConfirmed:
				mail.EnableAccountConfirmed(payload, isEmailAddressValid)
				break
			case queue.MessageMailNewLogin:
				mail.NewLogin(payload, isEmailAddressValid)
				break
			default:
				if s.Functions.HandleMail == nil || !s.Functions.HandleMail(msgType, payload, isEmailAddressValid) {
					log.WithFields(logrus.Fields{
						"topic": queue.TopicMail,
						"type":  msgType,
					}).Warn("unknown kafka message")
				}
			}
		}})

	return consumers
}

// Start start the worker.
func (s *Server) Start() {
	// Warn about the undefined functions.
	s.warnUndefinedFunctions()

	// Load the templates.
	html.LoadTemplatesWithDelimiters(config.Main.TemplatesPath+"/html", "[[", "]]", nil)
	text.LoadTemplates(config.Main.TemplatesPath + "/text")

	// Initialize the configuration of the HTML routes.
	weburl.Initialize(s.RoutesFile, config.Main.Hosts.Web)

	// Create the common vars for the emails.
	commonHTMLVars := make(map[string]map[string]interface{})
	commonHTMLVars["StaticHost"] = make(map[string]interface{})
	commonHTMLVars["HomeURL"] = make(map[string]interface{})
	commonHTMLVars["FacebookURL"] = make(map[string]interface{})
	commonHTMLVars["TwitterURL"] = make(map[string]interface{})

	commonTextVars := make(map[string]map[string]interface{})
	commonTextVars["StaticHost"] = make(map[string]interface{})
	commonTextVars["HomeURL"] = make(map[string]interface{})
	commonTextVars["FacebookURL"] = make(map[string]interface{})
	commonTextVars["TwitterURL"] = make(map[string]interface{})

	for _, locale := range config.Main.SupportedLocales {
		commonHTMLVars["StaticHost"][locale] = config.Main.Hosts.Static
		commonTextVars["StaticHost"][locale] = config.Main.Hosts.Static

		commonHTMLVars["HomeURL"][locale] = fmt.Sprintf("%s/%s/", config.Main.Hosts.Web, locale)
		commonTextVars["HomeURL"][locale] = fmt.Sprintf("%s/%s/", config.Main.Hosts.Web, locale)

		commonHTMLVars["FacebookURL"][locale] = config.Main.SocialNetworks.Facebook
		commonTextVars["FacebookURL"][locale] = config.Main.SocialNetworks.Facebook

		commonHTMLVars["TwitterURL"][locale] = config.Main.SocialNetworks.Twitter
		commonTextVars["TwitterURL"][locale] = config.Main.SocialNetworks.Twitter
	}

	// Set up the SMTP connection.
	mail.Initialize("", config.Main.SMTP.Username, config.Main.SMTP.Password, config.Main.SMTP.Host, config.Main.SMTP.Port, commonHTMLVars, commonTextVars)

	// Run the check up tasks.
	if s.Functions.StartUpTasks != nil && config.Main.RunStartUpTasks {
		log.Info("check up started")
		s.Functions.StartUpTasks(config.Main.Environment, config.Main.Version)
		log.Info("check up done")
	}

	// Load the recurring tasks.
	recurring.LoadScheduledTasks(config.Main.RecurringTasks)
	for _, task := range config.Main.RecurringTasks {
		switch task.Name {
		// Delete the expired auth tokens.
		case recurring.DeleteExpiredTokens:
			if s.Functions.DeleteExpiredTokens == nil {
				break
			}
			recurring.SetFunc(task.Name, func(taskName string) {
				authtokens.DeleteExpired(taskName, s.Functions.DeleteExpiredTokens)
			})
			break
		// Delete/update the data associated to deleted accounts.
		case recurring.SanitizeAccounts:
			if s.Functions.SanitizeAccounts == nil {
				break
			}
			recurring.SetFunc(task.Name, func(taskName string) {
				account.Sanitize(taskName, s.Functions.SanitizeAccounts)
			})
			break
		// Delete the never used accounts.
		case recurring.DeleteNeverUsedAccounts:
			if !config.Main.Account.EmailAddressMustBeConfirmed || s.Functions.GetNeverUsedAccounts == nil || s.Functions.IsUserEmpty == nil || s.Functions.DeleteUserByID == nil {
				break
			}
			recurring.SetFunc(task.Name, func(taskName string) {
				account.DeleteNeverUsed(taskName, config.Main.Account.DeleteNeverUsedAfter, s.Functions.GetNeverUsedAccounts, s.Functions.IsUserEmpty, s.Functions.DeleteUserByID)
			})
			break
		// Delete the inactive accounts.
		case recurring.DeleteInactiveAccounts:
			if s.Functions.GetInactiveAccounts == nil || s.Functions.DeleteUserByID == nil {
				break
			}
			recurring.SetFunc(task.Name, func(taskName string) {
				account.DeleteInactive(taskName, config.Main.Account.DeleteInactiveAfter, s.Functions.GetInactiveAccounts, s.Functions.DeleteUserByID)
			})
			break
		// Send again the registration emails.
		case recurring.SendRegistrationMails:
			break
		// Generate the XML sitemap files.
		case recurring.GenerateSitemaps:
			if s.Functions.GetLocalizedRouteVariables == nil {
				break
			}
			recurring.SetFunc(task.Name, func(taskName string) {
				sitemap.Generate(taskName, s.Functions.GetLocalizedRouteVariables)
			})
			break
		default:
		}
	}

	if s.Functions.SetFuncOfRecurringTasks != nil {
		s.Functions.SetFuncOfRecurringTasks(recurring.SetFunc)
	}

	// Start the recurring tasks.
	recurring.StartScheduledTasks()

	log.Info("started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)
	<-signals

	log.Info("shutdown")

	os.Exit(1)
}

// NewServer : returns a new struct 'Server'.
func NewServer() *Server {
	return &Server{}
}

// Helping functions.
func (s *Server) warnUndefinedFunctions() {
	if s.Functions.DeleteExpiredTokens == nil {
		log.Warn("The function 'DeleteExpiredTokens' is undefined.")
	}
	if s.Functions.DeleteUserByID == nil {
		log.Warn("The function 'DeleteUserByID' is undefined.")
	}
	if s.Functions.DisableUserByID == nil {
		log.Warn("The function 'DisableUserByID' is undefined.")
	}
	if s.Functions.EnableUserByID == nil {
		log.Warn("The function 'EnableUserByID' is undefined.")
	}
	if s.Functions.UpdateUserSocialNetworksByID == nil {
		log.Warn("The function 'UpdateUserSocialNetworksByID' is undefined.")
	}
	if s.Functions.GetNotificationTargetsByUserAndType == nil {
		log.Warn("The function 'GetNotificationTargetsByUserAndType' is undefined.")
	}
	if s.Functions.SaveIPAddressDetails == nil {
		log.Warn("The function 'SaveIPAddressDetails' is undefined.")
	}
	if s.Functions.SetFuncOfRecurringTasks == nil {
		log.Warn("The function 'SetFuncOfRecurringTasks' is undefined.")
	}
	if s.Functions.StartUpTasks == nil {
		log.Warn("The function 'StartUpTasks' is undefined.")
	}
	if s.Functions.HandleTask == nil {
		log.Warn("The function 'HandleTask' is undefined.")
	}
	if s.Functions.HandleUserNotification == nil {
		log.Warn("The function 'HandleUserNotification' is undefined.")
	}
	if s.Functions.HandleMail == nil {
		log.Warn("The function 'HandleMail' is undefined.")
	}
	if s.Functions.HandleNotificationForUsersWithAbility == nil {
		log.Warn("The function 'HandleNotificationForUsersWithAbility' is undefined.")
	}
	if s.Functions.IsEmailAddressValid == nil {
		log.Warn("The function 'IsEmailAddressValid' is undefined.")
	}
	if s.Functions.GetLocalizedRouteVariables == nil {
		log.Warn("The function 'GetLocalizedRouteVariables' is undefined.")
	}
}
