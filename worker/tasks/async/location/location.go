// Package location retrieves and saves the location of the IP addresses, from the geoip API
// (https://github.com/cygy/geoip).
package location

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"
)

const (
	// defaultTimeout is the maximum duration of a request to the API when the configuration sets none.
	defaultTimeout = 5 * time.Second

	// maxResponseSize is the maximum size of the response of the API.
	maxResponseSize = 1 << 20
)

// DetailsFunc returns the location of an IP address. It returns nil details, and no error, if the location is
// unknown or can not be retrieved, and an error only if asking again later can succeed.
type DetailsFunc func(IPAddress string) (*IPAddressDetails, error)

// IPAddressDetails is the location of an IP address, as the geoip API returns it.
type IPAddressDetails struct {
	Region      string `json:"region"`
	City        string `json:"city"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
}

// GetIPInfo decodes the payload of a location task and calls saveIPAddressDetailsFunc, which saves the location
// returned by the function it receives, which uses the geoip API of the configuration. Nothing happens if the function
// is nil or if the host of the API is empty (the location is optional): the database is not read either. It returns
// the error of the function, or a permanent error, see broker.Permanent, if the payload can not be decoded.
func GetIPInfo(payload []byte, geoIP config.GeoIPConfiguration, saveIPAddressDetailsFunc func(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc DetailsFunc) error) error {
	if saveIPAddressDetailsFunc == nil || geoIP.Host == "" {
		return nil
	}

	p := queue.TaskIPLocation{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	return saveIPAddressDetailsFunc(context.Background(), p.IPAddress, p.TokenID, func(IPAddress string) (*IPAddressDetails, error) {
		return GetIPAddressDetails(geoIP, IPAddress), nil
	})
}

// GetIPAddressDetails returns the location of an IP address from the geoip API of the configuration, or nil if it is
// unknown or can not be retrieved. The service is optional: nil is returned, without any request, if its host is
// empty. A failure (the API is unreachable, answers an unexpected status, or its response can not be read) is logged
// as an error and gives nil, without retry: the token keeps no location. An unknown address (404) is not an error.
func GetIPAddressDetails(geoIP config.GeoIPConfiguration, IPAddress string) *IPAddressDetails {
	if geoIP.Host == "" {
		return nil
	}

	timeout := geoIP.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	endpoint := strings.TrimSuffix(geoIP.Host, "/") + "/v1/ip/" + url.PathEscape(IPAddress)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		logError(IPAddress, err)
		return nil
	}
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		logError(IPAddress, err)
		return nil
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			logError(IPAddress, err)
		}
	}()

	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	if response.StatusCode != http.StatusOK {
		logError(IPAddress, fmt.Errorf("unexpected status %d from the API", response.StatusCode))
		return nil
	}

	content, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize))
	if err != nil {
		logError(IPAddress, err)
		return nil
	}

	details := IPAddressDetails{}
	if err := json.Unmarshal(content, &details); err != nil {
		logError(IPAddress, err)
		return nil
	}

	if details.CountryCode == "" && details.CountryName == "" && details.Region == "" {
		logError(IPAddress, errors.New("get no location from the API"))
		return nil
	}

	return &details
}

// logError logs that the location of an IP address can not be retrieved.
func logError(IPAddress string, err error) {
	log.Error("unable to get the location of the IP address",
		"IP", IPAddress,
		"error", err.Error(),
	)
}
