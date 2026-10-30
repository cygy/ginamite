// Package robots holds the content of the robots.txt file served by the web server.
package robots

import (
	"fmt"
	"os"
)

// Content is the content of the robots file, set by Initialize.
var Content string

// Initialize loads the robots file into Content.
func Initialize(filepath string) error {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("unable to load the robots file: %w", err)
	}

	Content = string(bytes)

	return nil
}
