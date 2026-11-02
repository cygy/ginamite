package kafka

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/metrics"
)

func TestClient_Collect(t *testing.T) {
	client := NewClient()
	client.stats.countReceived()
	client.stats.countReceived()
	client.stats.countHandled()
	client.stats.countDeadLettered()
	client.stats.setLag("app_users", 2, 7)

	registry := metrics.NewRegistry()
	registry.Register(client)

	recorder := httptest.NewRecorder()
	registry.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	got := recorder.Body.String()

	for _, want := range []string{
		"kafka_messages_received_total 2\n",
		"kafka_messages_handled_total 1\n",
		"kafka_messages_dead_lettered_total 1\n",
		"kafka_messages_retried_total 0\n",
		"kafka_outbox_messages 0\n",
		`kafka_consumer_lag_messages{topic="app_users",partition="2"} 7` + "\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("metrics lack %q:\n%s", want, got)
		}
	}
}
