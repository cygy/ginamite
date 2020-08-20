package api

// Initialize : initializes the client to the internal API.
func Initialize(host, version string, port int, timeout, retryCount int, debug bool) {
	Main = NewClient(host, version, port, timeout, retryCount, debug)
}

// AddResponseHandlers : adds some response handlers to the current response handlers.
func AddResponseHandlers(handlers ...ResponseHandlerFunc) {
	Main.AddResponseHandlers(handlers...)
}
