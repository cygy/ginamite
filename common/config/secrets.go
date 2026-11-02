package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Tags of the YAML values read from outside of the configuration file, so that it holds no secret:
//
//	password: !file /run/secrets/db_password
//	secret: !env JWT_SECRET
const (
	fileTag = "!file"
	envTag  = "!env"
)

// placeholderSecretPrefix starts the values of the example configurations, which are not secrets.
const placeholderSecretPrefix = "change-me"

// resolveExternalValues replaces the scalars tagged !file by the content of the file, without its trailing line
// break, and the ones tagged !env by the value of the environment variable. The result is read as if it was written
// in the file. An error names the key and the file or the variable, never the value.
func resolveExternalValues(node *yaml.Node, key string) error {
	if node.Kind == yaml.ScalarNode && (node.Tag == fileTag || node.Tag == envTag) {
		value, err := externalValue(node)
		if err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}

		node.Value, node.Tag, node.Style = value, "", 0

		return nil
	}

	if node.Kind == yaml.DocumentNode {
		for _, child := range node.Content {
			if err := resolveExternalValues(child, key); err != nil {
				return err
			}
		}

		return nil
	}

	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			child := key + "." + node.Content[i].Value
			if key == "" {
				child = node.Content[i].Value
			}

			if err := resolveExternalValues(node.Content[i+1], child); err != nil {
				return err
			}
		}

		return nil
	}

	for i, child := range node.Content {
		if err := resolveExternalValues(child, fmt.Sprintf("%s[%d]", key, i)); err != nil {
			return err
		}
	}

	return nil
}

// externalValue returns the value of a scalar tagged !file or !env.
func externalValue(node *yaml.Node) (string, error) {
	if node.Value == "" {
		return "", fmt.Errorf("the tag %s needs a value", node.Tag)
	}

	if node.Tag == envTag {
		value, ok := os.LookupEnv(node.Value)
		if !ok {
			return "", fmt.Errorf("the environment variable %s is not set", node.Value)
		}

		return value, nil
	}

	content, err := os.ReadFile(node.Value)
	if err != nil {
		return "", fmt.Errorf("unable to read the file %s: %w", node.Value, err)
	}

	return strings.TrimRight(string(content), "\r\n"), nil
}

// unmarshal decodes the YAML source into each of the outs, after resolving the values tagged !file and !env.
func unmarshal(source []byte, outs ...any) error {
	var document yaml.Node
	if err := yaml.Unmarshal(source, &document); err != nil {
		return err
	}

	if document.Kind == 0 {
		return nil // Empty file.
	}

	if err := resolveExternalValues(&document, ""); err != nil {
		return err
	}

	for _, out := range outs {
		if err := document.Decode(out); err != nil {
			return err
		}
	}

	return nil
}

// errPlaceholderSecret is returned when the secret of the tokens is the one of the examples.
var errPlaceholderSecret = errors.New("jwt.secret is still the placeholder of the examples: set a secret, for instance with !file")
