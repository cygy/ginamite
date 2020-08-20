package kafka

import (
	"os"

	"github.com/Shopify/sarama"
)

// Client : a kafka client. Could be a message producer.
type Client struct {
	HasProducer     bool
	Producer        sarama.AsyncProducer
	ProducerSignals chan os.Signal
	TopicsPrefix    string
}

// NewClient : initialize a new Client struct.
func NewClient() (client *Client) {
	client = &Client{}
	client.ProducerSignals = make(chan os.Signal, 1)
	return
}
