// Package metrics exposes the counters and the gauges of the servers in the text format of Prometheus, without
// depending on its client library.
package metrics

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// shutdownTimeout is the time the metrics server waits for the running requests when it is stopped.
const shutdownTimeout = 5 * time.Second

// Collector writes its metrics when the registry is scraped.
type Collector interface {
	Collect(w *Writer)
}

// Writer writes the metrics. Each metric is written with its help and type once, then with its samples, so a metric
// must be written by a single collector, with all its samples in a row.
type Writer struct {
	out   io.Writer
	typed map[string]bool
}

// Counter writes a sample of a counter. The labels are pairs of a name and a value.
func (w *Writer) Counter(name, help string, value float64, labels ...string) {
	w.sample("counter", name, help, value, labels)
}

// Gauge writes a sample of a gauge. The labels are pairs of a name and a value.
func (w *Writer) Gauge(name, help string, value float64, labels ...string) {
	w.sample("gauge", name, help, value, labels)
}

func (w *Writer) sample(kind, name, help string, value float64, labels []string) {
	if !w.typed[name] {
		w.typed[name] = true
		_, _ = fmt.Fprintf(w.out, "# HELP %s %s\n# TYPE %s %s\n", name, help, name, kind)
	}

	_, _ = fmt.Fprintf(w.out, "%s%s %s\n", name, formatLabels(labels), strconv.FormatFloat(value, 'g', -1, 64))
}

// formatLabels returns the labels as {name="value",...}, with the values escaped, or "" if there is none. A name
// without a value is ignored.
func formatLabels(labels []string) string {
	if len(labels) < 2 {
		return ""
	}

	var b strings.Builder
	b.WriteByte('{')
	for i := 0; i+1 < len(labels); i += 2 {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(labels[i])
		b.WriteString(`="`)
		b.WriteString(strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(labels[i+1]))
		b.WriteByte('"')
	}
	b.WriteByte('}')

	return b.String()
}

// Registry is the set of the collectors which are scraped together.
type Registry struct {
	mu         sync.Mutex
	collectors []Collector
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a collector. It ignores nil.
func (r *Registry) Register(c Collector) {
	if c == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.collectors = append(r.collectors, c)
}

// ServeHTTP writes the metrics of the collectors.
func (r *Registry) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	r.mu.Lock()
	collectors := slices.Clone(r.collectors)
	r.mu.Unlock()

	rw.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	w := &Writer{out: rw, typed: map[string]bool{}}
	for _, c := range collectors {
		c.Collect(w)
	}
}

// Serve serves the metrics, on /metrics, at the address (host:port) until ctx is done. It returns once the server is
// stopped, or if it can not listen.
func (r *Registry) Serve(ctx context.Context, address string) error {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("unable to listen to %s: %w", address, err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", r)
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
