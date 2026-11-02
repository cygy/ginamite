package kafka

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/config"
)

// writeCertificate writes a self-signed certificate and its key in the directory, and returns their paths.
func writeCertificate(t *testing.T, dir string) (certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "kafka"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	certFile, keyFile = filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}), 0o600); err != nil {
		t.Fatal(err)
	}

	return certFile, keyFile
}

func TestTLSSettings_TLSConfig(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := writeCertificate(t, dir)

	t.Run("disabled", func(t *testing.T) {
		config, err := TLSSettings{CAFile: "ignored"}.TLSConfig()
		if config != nil || err != nil {
			t.Errorf("config = %v, error = %v", config, err)
		}
	})

	t.Run("system authorities", func(t *testing.T) {
		config, err := TLSSettings{Enabled: true}.TLSConfig()
		if err != nil || config == nil || config.RootCAs != nil || len(config.Certificates) != 0 || config.InsecureSkipVerify {
			t.Errorf("config = %+v, error = %v", config, err)
		}
	})

	t.Run("authority and client certificate", func(t *testing.T) {
		config, err := TLSSettings{Enabled: true, CAFile: certFile, CertFile: certFile, KeyFile: keyFile, InsecureSkipVerify: true}.TLSConfig()

		if err != nil || config.RootCAs == nil || len(config.Certificates) != 1 || !config.InsecureSkipVerify {
			t.Errorf("config = %+v, error = %v", config, err)
		}
	})

	t.Run("failures", func(t *testing.T) {
		notPEM := filepath.Join(dir, "empty.pem")
		if err := os.WriteFile(notPEM, []byte("nope"), 0o600); err != nil {
			t.Fatal(err)
		}

		tests := map[string]TLSSettings{
			"missing authority":     {Enabled: true, CAFile: filepath.Join(dir, "missing.pem")},
			"authority without PEM": {Enabled: true, CAFile: notPEM},
			"missing key":           {Enabled: true, CertFile: certFile},
			"missing certificate":   {Enabled: true, KeyFile: keyFile},
			"invalid certificate":   {Enabled: true, CertFile: notPEM, KeyFile: keyFile},
		}
		for name, c := range tests {
			if config, err := c.TLSConfig(); err == nil || config != nil {
				t.Errorf("%s: config = %v, error = %v", name, config, err)
			}
		}
	})
}

func TestSettingsFrom(t *testing.T) {
	load := func(t *testing.T, content string) *config.Configuration {
		t.Helper()

		file := filepath.Join(t.TempDir(), "server.yaml")
		if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		c := config.NewConfiguration()
		if err := c.Initialize(file, "test", "final"); err != nil {
			t.Fatal(err)
		}

		return c
	}

	t.Run("defaults", func(t *testing.T) {
		settings, err := settingsFrom(load(t, "server: {port: 80}\n"))

		if err != nil || settings != DefaultSettings() {
			t.Errorf("settings = %+v, error = %v, want the defaults", settings, err)
		}
		if topics := settings.Topics; topics.Create || topics.Partitions != 1 || topics.ReplicationFactor != 1 || topics.RetentionHours != 168 || topics.DeadLetterRetentionHours != 720 {
			t.Errorf("topics = %+v, want the defaults", topics)
		}
		if settings.MaxAttempts != 3 || settings.RetryDelay != 1000 {
			t.Errorf("attempts = %d, delay = %d, want the defaults", settings.MaxAttempts, settings.RetryDelay)
		}
	})

	t.Run("file merged over the defaults", func(t *testing.T) {
		settings, err := settingsFrom(load(t, `
kafka:
  max_attempts: 5
  retry_delay: 250
  version: 3.6.0
  topics_prefix: app_
  tls: {enabled: true, ca_file: ca.pem}
  sasl: {mechanism: SCRAM-SHA-256, username: user, password: secret}
  topics: {create: true, partitions: 6, replication_factor: 3, retention_hours: 24}
  stats_interval: 30
`))
		if err != nil {
			t.Fatal(err)
		}

		want := Settings{
			Version:       "3.6.0",
			TLS:           TLSSettings{Enabled: true, CAFile: "ca.pem"},
			SASL:          SASLSettings{Mechanism: "SCRAM-SHA-256", Username: "user", Password: "secret"},
			Topics:        TopicsSettings{Create: true, Partitions: 6, ReplicationFactor: 3, RetentionHours: 24, DeadLetterRetentionHours: 720},
			TopicsPrefix:  "app_",
			MaxAttempts:   5,
			RetryDelay:    250,
			StatsInterval: 30,
		}
		if settings != want {
			t.Errorf("settings = %+v, want %+v", settings, want)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := settingsFrom(load(t, "kafka: {max_attempts: many}\n")); err == nil {
			t.Error("error = nil")
		}
	})
}
