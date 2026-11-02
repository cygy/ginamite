// Package ctxkey stores and reads the services shared by the handlers in the gin context: the configuration, the
// translation engine, the queue, etc. Each service has a typed key, saved by a middleware and read by the handlers.
//
//	router.Use(ctxkey.Config.Inject(configuration))
//
//	func handler(c *gin.Context) {
//		configuration := ctxkey.Config.Get(c)
//	}
package ctxkey

import (
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

// The keys of the services saved by the servers.
var (
	// Config is the configuration of the application.
	Config = New[*config.Configuration]("config")
	// Localization is the translation engine.
	Localization = New[*localization.Engine]("localization")
	// Notifications is the registry of the notifications.
	Notifications = New[*notifications.Registry]("notifications")
	// Queue is the queue sending the messages to the other servers.
	Queue = New[*queue.Queue]("queue")
	// Timezones are the timezones of the countries.
	Timezones = New[*timezone.Configuration]("timezones")
)

// Key is the typed key of a service in the gin context.
type Key[T any] struct {
	name string
}

// New returns the key of a service. The name is only used for the key in the context and in the panic message.
func New[T any](name string) Key[T] {
	return Key[T]{name: name}
}

// Inject returns a middleware saving the service in the gin context, see Get.
func (k Key[T]) Inject(service T) gin.HandlerFunc {
	return func(c *gin.Context) {
		k.Set(c, service)
		c.Next()
	}
}

// Set saves the service in the gin context, like Inject does. It is meant for the tests.
func (k Key[T]) Set(c *gin.Context, service T) {
	c.Set(k.contextKey(), service)
}

// Get returns the service saved in the gin context by Inject.
// It panics if there is none: the handlers can not work without their services.
func (k Key[T]) Get(c *gin.Context) T {
	value, ok := c.Get(k.contextKey())
	if !ok {
		panic("ctxkey: no " + k.name + " in the context, add the middleware Inject")
	}

	return value.(T)
}

func (k Key[T]) contextKey() string {
	return "_" + k.name
}
