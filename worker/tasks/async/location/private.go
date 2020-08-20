package location

import (
	"github.com/cygy/ginamite/common/log"
	"github.com/sirupsen/logrus"
)

// Helping function.
func logError(IPAddress string, err error) {
	log.WithFields(logrus.Fields{
		"IP":    IPAddress,
		"error": err.Error(),
	}).Error("unable to get or save the location of the IP address")
}
