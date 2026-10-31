// Package robots loads the content of the robots.txt file served by the web server.
package robots

import (
	"fmt"
	"os"
)

// Load returns the content of the robots file.
func Load(filepath string) (string, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("unable to load the robots file: %w", err)
	}

	return string(bytes), nil
}
