package config

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

func TestKafkaTLSConfiguration_TLSConfig(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := writeCertificate(t, dir)

	t.Run("disabled", func(t *testing.T) {
		config, err := KafkaTLSConfiguration{CAFile: "ignored"}.TLSConfig()
		if config != nil || err != nil {
			t.Errorf("config = %v, error = %v", config, err)
		}
	})

	t.Run("system authorities", func(t *testing.T) {
		config, err := KafkaTLSConfiguration{Enabled: true}.TLSConfig()
		if err != nil || config == nil || config.RootCAs != nil || len(config.Certificates) != 0 || config.InsecureSkipVerify {
			t.Errorf("config = %+v, error = %v", config, err)
		}
	})

	t.Run("authority and client certificate", func(t *testing.T) {
		config, err := KafkaTLSConfiguration{Enabled: true, CAFile: certFile, CertFile: certFile, KeyFile: keyFile, InsecureSkipVerify: true}.TLSConfig()

		if err != nil || config.RootCAs == nil || len(config.Certificates) != 1 || !config.InsecureSkipVerify {
			t.Errorf("config = %+v, error = %v", config, err)
		}
	})

	t.Run("failures", func(t *testing.T) {
		notPEM := filepath.Join(dir, "empty.pem")
		if err := os.WriteFile(notPEM, []byte("nope"), 0o600); err != nil {
			t.Fatal(err)
		}

		tests := map[string]KafkaTLSConfiguration{
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
