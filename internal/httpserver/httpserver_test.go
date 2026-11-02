package httpserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func freeAddress(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()

	return l.Addr().String()
}

func TestServe_StopsGracefully(t *testing.T) {
	address := freeAddress(t)
	started := make(chan struct{})
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "done")
	})

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- serve(ctx, newServer(address, handler), func(s *http.Server) error { return s.ListenAndServe() })
	}()

	body := make(chan string, 1)
	go func() {
		var response *http.Response
		var err error
		for range 100 {
			if response, err = http.Get("http://" + address); err == nil { //nolint:noctx // a test against a local server
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if err != nil {
			body <- err.Error()
			return
		}
		content, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		body <- string(content)
	}()

	<-started
	cancel() // the request is running: the server waits for it
	select {
	case err := <-result:
		t.Fatalf("serve returned %v before the request ended", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	if got := <-body; got != "done" {
		t.Errorf("body = %q", got)
	}
	if err := <-result; err != nil {
		t.Errorf("serve() = %v", err)
	}
}

func TestServe_ListenError(t *testing.T) {
	if err := ListenAndServe(t.Context(), "invalid address", http.NotFoundHandler()); err == nil {
		t.Error("an invalid address must fail")
	}
	if err := ListenAndServeTLS(t.Context(), "invalid address", "cert", "key", http.NotFoundHandler()); err == nil {
		t.Error("an invalid address must fail")
	}
}

func TestListenAndServe_Context(t *testing.T) {
	address := freeAddress(t)
	ctx, cancel := context.WithCancel(t.Context())
	result := make(chan error, 1)
	go func() { result <- ListenAndServe(ctx, address, http.NotFoundHandler()) }()

	for range 100 {
		if c, err := net.Dial("tcp", address); err == nil {
			_ = c.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-result:
		if err != nil {
			t.Errorf("ListenAndServe() = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the server did not stop once the context is done")
	}
}
