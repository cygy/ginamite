package validator

import resty "github.com/go-resty/resty/v2"

type facebookTokenValidator struct {
	client        *resty.Client
	applicationID string
}
