package weburl

// GetURL : returns the full URL of a web route.
func GetURL(key, locale string) string {
	return Main.GetURL(key, locale)
}

// Initialize : initializes the configuration of the routes.
func Initialize(file, webHost string) bool {
	Main = NewConfiguration()
	return Main.Initialize(file, webHost)
}
