// Package httpserver runs an HTTP server which stops gracefully on SIGINT and SIGTERM, with timeouts against the slow
// clients.
package httpserver

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const (
	// shutdownTimeout is the time the running requests have to end when the server is stopped.
	shutdownTimeout = 30 * time.Second

	readHeaderTimeout = 10 * time.Second
	readTimeout       = 60 * time.Second
	writeTimeout      = 120 * time.Second
	idleTimeout       = 120 * time.Second
)

// ListenAndServe serves the handler at the address until SIGINT or SIGTERM, then waits for the running requests. It
// returns nil if the server stopped that way.
func ListenAndServe(address string, handler http.Handler) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return serve(ctx, newServer(address, handler), func(s *http.Server) error { return s.ListenAndServe() })
}

// ListenAndServeTLS is like ListenAndServe with TLS, using the certificate and the key files.
func ListenAndServeTLS(address, certPath, keyPath string, handler http.Handler) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return serve(ctx, newServer(address, handler), func(s *http.Server) error { return s.ListenAndServeTLS(certPath, keyPath) })
}

func newServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// serve runs listen, and shuts the server down when ctx is done.
func serve(ctx context.Context, server *http.Server, listen func(*http.Server) error) error {
	failed := make(chan error, 1)
	go func() { failed <- listen(server) }()

	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-failed; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
