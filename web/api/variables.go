package api

var (
	// Main : main API client of the application.
	Main *Client

	// MainCache : main cache for the results of some API calls.
	MainCache = NewCache()
)
