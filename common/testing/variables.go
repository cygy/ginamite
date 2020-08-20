package testing

import "os"

var (
	apiAddress = os.Getenv("API_HOST") + ":" + os.Getenv("API_PORT")

	// APIVersion version of the API to test
	APIVersion = ""
)
