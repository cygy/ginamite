package config

import (
	"errors"

	"github.com/cygy/ginamite/common/yaml"
)

// placeholderSecretPrefix starts the values of the example configurations, which are not secrets.
const placeholderSecretPrefix = "change-me"

// unmarshal decodes the YAML source into each of the outs, after resolving the values tagged !file and !env.
func unmarshal(source []byte, outs ...any) error { return yaml.Unmarshal(source, outs...) }

// errPlaceholderSecret is returned when the secret of the tokens is the one of the examples.
var errPlaceholderSecret = errors.New("jwt.secret is still the placeholder of the examples: set a secret, for instance with !file")

// removedRateLimitsSection is the section of the limits of the requests before they moved under api_rate_limiter.
const removedRateLimitsSection = "rate_limits"

// errRateLimitsRemoved is returned when the configuration still has the section rate_limits, which would otherwise be
// ignored silently, and the default limits applied instead.
var errRateLimitsRemoved = errors.New("the section rate_limits is removed: move it to api_rate_limiter.limits")
