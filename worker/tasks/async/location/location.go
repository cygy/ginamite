// Package location retrieves and saves the location of the IP addresses.
package location

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"
)

const (
	// requestTimeout is the maximum duration of a request to the API.
	requestTimeout = 10 * time.Second

	// maxResponseSize is the maximum size of the response of the API.
	maxResponseSize = 1 << 20
)

// DetailsFunc returns the location of an IP address. It returns nil details, and no error, if the location is
// unknown and asking again would not change it, and an error if the API failed and asking again later can succeed.
type DetailsFunc func(IPAddress string) (*IPAddressDetails, error)

// IPAddressDetails is the location of an IP address.
type IPAddressDetails struct {
	Region      string `json:"region_name"`
	City        string `json:"city"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
}

// GetIPInfo decodes the payload of a location task and calls saveIPAddressDetailsFunc, which saves
// the location returned by the function it receives, which uses the ipstack API of the configuration.
// Nothing happens if the function is nil. It returns the error of the function, which includes the temporary
// failures of the API, or a permanent error, see broker.Permanent, if the payload can not be decoded.
func GetIPInfo(payload []byte, ipStack config.IPStackConfiguration, saveIPAddressDetailsFunc func(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc DetailsFunc) error) error {
	if saveIPAddressDetailsFunc == nil {
		return nil
	}

	p := queue.TaskIPLocation{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	return saveIPAddressDetailsFunc(context.Background(), p.IPAddress, p.TokenID, func(IPAddress string) (*IPAddressDetails, error) {
		return GetIPAddressDetails(ipStack, IPAddress)
	})
}

// GetIPAddressDetails returns the location of an IP address from the ipstack API of the configuration.
// A failure which can be temporary (the API is unreachable, answers too many requests or a server error, or its
// response can not be read) is returned as an error, without the key of the API. A response which would not change
// by asking again (unknown location, client error, invalid content) is logged and gives nil details and no error.
func GetIPAddressDetails(ipStack config.IPStackConfiguration, IPAddress string) (*IPAddressDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	endpoint := fmt.Sprintf("%s/%s?access_key=%s", ipStack.Host, url.PathEscape(IPAddress), url.QueryEscape(ipStack.Key))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, withoutURL(err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, withoutURL(err)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			logError(IPAddress, err)
		}
	}()

	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("unexpected status %d from the API", response.StatusCode)
	}
	if response.StatusCode != http.StatusOK {
		logError(IPAddress, fmt.Errorf("unexpected status %d from the API", response.StatusCode))
		return nil, nil
	}

	content, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize))
	if err != nil {
		return nil, withoutURL(err)
	}

	details := IPAddressDetails{}
	if err := json.Unmarshal(content, &details); err != nil {
		logError(IPAddress, err)
		return nil, nil
	}

	if len(details.Region) == 0 && len(details.CountryName) == 0 && len(details.CountryCode) == 0 {
		logError(IPAddress, errors.New("get no location from the API"))
		return nil, nil
	}

	return &details, nil
}

// withoutURL removes the URL of a failed request from the error: it contains the key of the API.
func withoutURL(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err
	}

	return err
}

// logError logs that the location of an IP address can not be retrieved, and will not be by trying again.
func logError(IPAddress string, err error) {
	log.Error("unable to get or save the location of the IP address",
		"IP", IPAddress,
		"error", withoutURL(err).Error(),
	)
}
