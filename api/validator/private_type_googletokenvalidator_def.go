package validator

import resty "github.com/go-resty/resty/v2"

type googleTokenValidator struct {
	client   *resty.Client
	clientID string
}
