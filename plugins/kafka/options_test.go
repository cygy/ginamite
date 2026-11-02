package kafka

import (
	"crypto/tls"
	"testing"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

func TestOptions_NewConfig(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		config := Options{}.NewConfig()

		if err := config.Validate(); err != nil {
			t.Fatalf("the configuration is invalid: %v", err)
		}
		if config.Producer.RequiredAcks != sarama.WaitForAll || !config.Producer.Idempotent || config.Net.MaxOpenRequests != 1 || config.Producer.Retry.Max == 0 {
			t.Errorf("the producer must wait for all the replicas and not duplicate: %+v", config.Producer)
		}
		if !config.Producer.Return.Successes || !config.Producer.Return.Errors || !config.Consumer.Return.Errors {
			t.Error("the successes and the errors must be returned")
		}
		if config.Consumer.Offsets.Initial != sarama.OffsetNewest {
			t.Errorf("initial offset = %d", config.Consumer.Offsets.Initial)
		}
		if config.Version != sarama.NewConfig().Version {
			t.Errorf("version = %v, want the one of the library", config.Version)
		}
		if config.Net.TLS.Enable || config.Net.SASL.Enable {
			t.Error("neither TLS nor SASL must be enabled")
		}
	})

	t.Run("version", func(t *testing.T) {
		if config := (Options{Version: sarama.V3_6_0_0}).NewConfig(); config.Version != sarama.V3_6_0_0 {
			t.Errorf("version = %v", config.Version)
		}
	})

	t.Run("TLS", func(t *testing.T) {
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
		config := Options{TLS: tlsConfig}.NewConfig()

		if !config.Net.TLS.Enable || config.Net.TLS.Config != tlsConfig {
			t.Errorf("TLS = %+v", config.Net.TLS)
		}
	})

	t.Run("SASL", func(t *testing.T) {
		for _, mechanism := range []string{SASLPlain, SASLSCRAMSHA256, SASLSCRAMSHA512} {
			config := Options{SASL: SASL{Mechanism: mechanism, Username: "user", Password: "secret"}}.NewConfig()

			if !config.Net.SASL.Enable || string(config.Net.SASL.Mechanism) != mechanism || config.Net.SASL.User != "user" || config.Net.SASL.Password != "secret" {
				t.Errorf("%s: SASL = %+v", mechanism, config.Net.SASL)
			}
			if (config.Net.SASL.SCRAMClientGeneratorFunc != nil) != (mechanism != SASLPlain) {
				t.Errorf("%s: SCRAM client generator defined = %v", mechanism, config.Net.SASL.SCRAMClientGeneratorFunc != nil)
			}
			if err := config.Validate(); err != nil {
				t.Errorf("%s: the configuration is invalid: %v", mechanism, err)
			}
		}
	})
}

func TestScramClient(t *testing.T) {
	for name, hash := range map[string]scram.HashGeneratorFcn{"SHA-256": scram.SHA256, "SHA-512": scram.SHA512} {
		t.Run(name, func(t *testing.T) {
			// The server knows the credentials of the user, not the password.
			reference, err := hash.NewClient("user", "secret", "")
			if err != nil {
				t.Fatal(err)
			}
			stored, err := reference.GetStoredCredentialsWithError(scram.KeyFactors{Salt: "salt", Iters: 4096})
			if err != nil {
				t.Fatal(err)
			}
			server, err := hash.NewServer(func(string) (scram.StoredCredentials, error) { return stored, nil })
			if err != nil {
				t.Fatal(err)
			}

			authenticate := func(password string) error {
				var client sarama.SCRAMClient = &scramClient{hash: hash}
				if err := client.Begin("user", password, ""); err != nil {
					return err
				}

				conversation := server.NewConversation()
				challenge := ""
				for !client.Done() {
					response, err := client.Step(challenge)
					if err != nil {
						return err
					}

					// The last message of the server is the last one of the exchange.
					if conversation.Done() {
						break
					}

					if challenge, err = conversation.Step(response); err != nil {
						return err
					}
				}

				if !conversation.Valid() {
					t.Error("the server did not accept the user")
				}

				return nil
			}

			if err := authenticate("secret"); err != nil {
				t.Errorf("authentication with the right password: %v", err)
			}
			if err := authenticate("wrong"); err == nil {
				t.Error("authentication with the wrong password must fail")
			}
		})
	}

	t.Run("invalid user name", func(t *testing.T) {
		var client sarama.SCRAMClient = &scramClient{hash: scram.SHA256}
		if err := client.Begin("user", "secret\x00", ""); err == nil {
			t.Error("Begin() must fail with an invalid password")
		}
	})
}
