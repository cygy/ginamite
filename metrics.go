package ginamite

import (
	"context"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/metrics"
)

// startMetrics serves the Prometheus metrics on the address of the configuration, if any: the requests of the API and
// the broker, if it is a metrics.Collector, such as Kafka. It returns the function which stops the server.
func (s *Server) startMetrics() (stop func()) {
	address := s.Config.Metrics.Address
	if address == "" {
		return func() {}
	}

	registry := metrics.NewRegistry()

	s.API.Requests = metrics.NewRequests()
	registry.Register(s.API.Requests)

	if collector, ok := s.Broker.(metrics.Collector); ok {
		registry.Register(collector)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)

		if err := registry.Serve(ctx, address); err != nil {
			log.Error("unable to serve the metrics", "error", err)
		}
	}()

	log.Info("the metrics are served on /metrics", "address", address)

	return func() {
		cancel()
		<-done
	}
}
