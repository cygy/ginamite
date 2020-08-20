package robots

import (
	"io/ioutil"

	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// Initialize : initialize the robots file.
func Initialize(filepath string) error {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.WithFields(logrus.Fields{
			"path":  filepath,
			"error": err.Error(),
		}).Panic("unable to load the robots file")
		return err
	}

	Content = string(bytes)

	return nil
}
