package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// KafkaConfiguration holds the settings of the connection to Kafka and of its topics. The address of the broker is the
// service "kafka".
type KafkaConfiguration struct {
	// Version is the version of the brokers, such as "2.8.0". It is the one of the client library if empty.
	Version string
	TLS     KafkaTLSConfiguration  `yaml:"tls"`
	SASL    KafkaSASLConfiguration `yaml:"sasl"`
	Topics  KafkaTopicsConfiguration

	// StatsInterval is the interval, in seconds, of the log of the statistics of the producer and of the consumers.
	// They are not logged if 0.
	StatsInterval int `yaml:"stats_interval"`
}

// KafkaTLSConfiguration holds the settings of the encrypted connection to Kafka.
type KafkaTLSConfiguration struct {
	Enabled bool
	// CAFile is the certificate authority of the brokers. The system ones are used if empty.
	CAFile string `yaml:"ca_file"`
	// CertFile and KeyFile are the certificate and the key of the client, for a mutual authentication.
	CertFile           string `yaml:"cert_file"`
	KeyFile            string `yaml:"key_file"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

// KafkaSASLConfiguration holds the settings of the authentication to Kafka. There is no authentication if the
// mechanism is empty.
type KafkaSASLConfiguration struct {
	// Mechanism is PLAIN, SCRAM-SHA-256 or SCRAM-SHA-512. Enable TLS to protect the password.
	Mechanism string
	Username  string
	Password  string
}

// KafkaTopicsConfiguration holds the settings of the creation of the topics, which is done when the servers start
// if Create is true. The topics which exist are not changed.
type KafkaTopicsConfiguration struct {
	Create bool
	// Partitions is the number of partitions of a topic: the workers of a group share the partitions, so there is no
	// use for more workers than partitions. The messages of a topic are only ordered within a partition.
	Partitions        int32
	ReplicationFactor int16 `yaml:"replication_factor"`
	// RetentionHours is how long the messages are kept. DeadLetterRetentionHours is the one of the dead letters.
	RetentionHours           int `yaml:"retention_hours"`
	DeadLetterRetentionHours int `yaml:"dead_letter_retention_hours"`
}

// TLSConfig returns the TLS configuration of the connection to Kafka, or nil if TLS is not enabled.
func (c KafkaTLSConfiguration) TLSConfig() (*tls.Config, error) {
	if !c.Enabled {
		return nil, nil
	}

	config := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: c.InsecureSkipVerify} //nolint:gosec // opt-in, for tests

	if c.CAFile != "" {
		pem, err := os.ReadFile(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read the certificate authority: %w", err)
		}

		config.RootCAs = x509.NewCertPool()
		if !config.RootCAs.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("no certificate in %s", c.CAFile)
		}
	}

	if c.CertFile != "" || c.KeyFile != "" {
		certificate, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("unable to load the certificate of the client: %w", err)
		}

		config.Certificates = []tls.Certificate{certificate}
	}

	return config, nil
}
