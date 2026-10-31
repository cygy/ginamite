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

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"

	"github.com/sirupsen/logrus"
)

const (
	// requestTimeout is the maximum duration of a request to the API.
	requestTimeout = 10 * time.Second

	// maxResponseSize is the maximum size of the response of the API.
	maxResponseSize = 1 << 20
)

// IPAddressDetails is the location of an IP address.
type IPAddressDetails struct {
	Region      string `json:"region_name"`
	City        string `json:"city"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
}

// GetIPInfo decodes the payload of a location task and calls saveIPAddressDetailsFunc, which saves
// the location returned by the function it receives, which uses the ipstack API of the configuration.
// Nothing happens if the function is nil.
func GetIPInfo(payload []byte, ipStack config.IPStackConfiguration, saveIPAddressDetailsFunc func(ctx context.Context, IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *IPAddressDetails)) {
	if saveIPAddressDetailsFunc == nil {
		return
	}

	p := queue.TaskIPLocation{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	saveIPAddressDetailsFunc(context.Background(), p.IPAddress, p.TokenID, func(IPAddress string) *IPAddressDetails {
		return GetIPAddressDetails(ipStack, IPAddress)
	})
}

// GetIPAddressDetails returns the location of an IP address from the ipstack API of the configuration.
// It logs the error, and returns nil, if the API fails or does not know the location.
func GetIPAddressDetails(ipStack config.IPStackConfiguration, IPAddress string) *IPAddressDetails {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	endpoint := fmt.Sprintf("%s/%s?access_key=%s", ipStack.Host, url.PathEscape(IPAddress), url.QueryEscape(ipStack.Key))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		logError(IPAddress, err)
		return nil
	}
	request.Header.Set("Content-Type", "application/json")

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

	if len(details.Region) == 0 && len(details.CountryName) == 0 && len(details.CountryCode) == 0 {
		logError(IPAddress, errors.New("get no location from the API"))
		return nil
	}

	return &details
}

// logError logs that the location of an IP address can not be retrieved.
// The URL of a failed request is removed from the error: it contains the key of the API.
func logError(IPAddress string, err error) {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}

	log.WithFields(logrus.Fields{
		"IP":    IPAddress,
		"error": err.Error(),
	}).Error("unable to get or save the location of the IP address")
}
