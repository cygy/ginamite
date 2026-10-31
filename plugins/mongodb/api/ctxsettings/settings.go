// Package ctxsettings stores and reads the settings of the MongoDB plugin in the gin context.
package ctxsettings

import (
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// settingsContextKey is the key of the settings of the plugin in the gin context.
const settingsContextKey = "_settings"

// InjectSettings returns a middleware saving the settings of the plugin in the gin context, see GetSettings.
func InjectSettings(settings model.Settings) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetSettings(c, settings)
		c.Next()
	}
}

// SetSettings saves the settings of the plugin in the gin context, like InjectSettings does. It is meant for the tests.
func SetSettings(c *gin.Context, settings model.Settings) {
	c.Set(settingsContextKey, settings)
}

// GetSettings returns the settings of the plugin saved in the gin context by InjectSettings.
// It panics if there are none: the functions of the plugin can not work without them.
func GetSettings(c *gin.Context) model.Settings {
	value, ok := c.Get(settingsContextKey)
	if !ok {
		panic("context: no settings of the plugin in the context, add the middleware InjectSettings")
	}

	return value.(model.Settings)
}
