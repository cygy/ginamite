package ginamite

import "github.com/cygy/ginamite/common/broker"

// startBroker starts the broker, set by a plugin, see UseBroker, with the consumers, and reports whether it did. The
// broker receives the deduplicator of the server first, if it ignores the messages delivered twice.
func (s *Server) startBroker(consumers []broker.MessageConsumer) bool {
	if s.Broker == nil {
		return false
	}

	if b, ok := s.Broker.(broker.DeduplicatorUser); ok {
		b.UseDeduplicator(s.Deduplicator)
	}
	s.Broker.Start(consumers)

	return true
}
